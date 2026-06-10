#!/bin/bash
echo "Creating Kafka topics..."

BROKER="kafka:29092"
KAFKA_BIN="/opt/kafka/bin"

$KAFKA_BIN/kafka-topics.sh --create --if-not-exists --bootstrap-server $BROKER \
  --partitions 3 --replication-factor 1 --topic server.created

$KAFKA_BIN/kafka-topics.sh --create --if-not-exists --bootstrap-server $BROKER \
  --partitions 3 --replication-factor 1 --topic server.updated

$KAFKA_BIN/kafka-topics.sh --create --if-not-exists --bootstrap-server $BROKER \
  --partitions 3 --replication-factor 1 --topic server.deleted

$KAFKA_BIN/kafka-topics.sh --create --if-not-exists --bootstrap-server $BROKER \
  --partitions 6 --replication-factor 1 --topic server.status.changed

$KAFKA_BIN/kafka-topics.sh --create --if-not-exists --bootstrap-server $BROKER \
  --partitions 3 --replication-factor 1 --topic server.health.batch

$KAFKA_BIN/kafka-topics.sh --create --if-not-exists --bootstrap-server $BROKER \
  --partitions 3 --replication-factor 1 --topic import.job.created

$KAFKA_BIN/kafka-topics.sh --create --if-not-exists --bootstrap-server $BROKER \
  --partitions 1 --replication-factor 1 --topic report.daily.trigger

echo "All topics created!"
$KAFKA_BIN/kafka-topics.sh --list --bootstrap-server $BROKER
