package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
)

func main() {

	if len(os.Args) < 4 {
		fmt.Fprintf(os.Stderr, "Usage: %s <broker> <group> <topics..>\n",
			os.Args[0])
		os.Exit(1)
	}

	broker := os.Args[1]
	group := os.Args[2]
	topics := os.Args[3:]

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":               	broker,
		"group.id":                        	group,
		"session.timeout.ms":              	10000,
		//"poll":								200,
		"go.events.channel.size":			8000,
		//"max.poll.records":					8000,
		"max.partition.fetch.bytes":		393216,
		"fetch.max.bytes":					1048576,
		"max.poll.interval.ms":				86400000,
		"go.events.channel.enable":        	true,
		"go.application.rebalance.enable": 	true,
		"enable.auto.commit":              	false,
		// Enable generation of PartitionEOF when the
		// end of a partition is reached.
		"enable.partition.eof": 			true,
		"auto.offset.reset":    			"earliest"})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create consumer: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("Created Consumer %v\n", c)

	err = c.SubscribeTopics(topics, nil)

	run := true
	worker := 0
	chs := map[int32]chan *kafka.Message{}

	for run == true {
		select {
		case sig := <-sigchan:
			fmt.Printf("Caught signal %v: terminating\n", sig)
			run = false

		case ev := <-c.Events():
			switch e := ev.(type) {
			case kafka.AssignedPartitions:
				log.Printf("%% %v\n", e)
				c.Assign(e.Partitions)
				// Create a channel for each partition
				// Start worker for each partition
				for _, p := range e.Partitions {
					if chs[p.Partition] == nil {
						log.Printf("Creating channel for partition %d", p.Partition)
						chs[p.Partition] = make(chan *kafka.Message, 2)

						log.Printf("Launching worker [%d]", worker)

						go func(worker int, ch chan *kafka.Message) {
							runWrk := true

							for runWrk == true {
								select {
								/*case sig := <-sigchan:
								log.Printf("Caught signal in worker %v: terminating\n", sig)
								run = false*/
								case m := <-ch:

									log.Printf("%% [worker-%d] Message on %s:\n%s\n", worker, m.TopicPartition, string(m.Value))

									// Empirical Observation: CommitMessage only commit offset if offset-1 has been committed previously.
									_, err := c.CommitMessage(m)

									if err != nil {
										log.Fatalf("%% Error %s commiting %v\n", err, e)
									} else {
										log.Printf("%% [worker-%d] Commiting on %s\n", worker, m.TopicPartition)
									}
								}
							}
						}(worker, chs[p.Partition])

						worker++
					}
				}
			case kafka.RevokedPartitions:
				log.Printf("%% %v\n", e)
				c.Unassign()
				/* We can NOT stop workers for revoked partitions, because when assignation changes,
				kafka first revoke all partitions always.
				So we leave the channel and it would be reused if partition is reassigned. */
			case *kafka.Message:
				ch := chs[e.TopicPartition.Partition]

				if ch == nil {
					log.Fatalf("Channel for message %s not found", e)
				} else {
					log.Printf("Sending kafka msg to channel for %d", e.TopicPartition.Partition)
					ch <- e
				}

			case kafka.PartitionEOF:
				log.Printf("%% Reached %v\n", e)
			case kafka.Error:
				// Errors should generally be considered as informational, the client will try to automatically recover
				fmt.Fprintf(os.Stderr, "%% Error: %v\n", e)
			}
		}
	}

	fmt.Printf("Closing consumer\n")

	c.Close()
}
