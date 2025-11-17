# HW10 Microservice Extravaganza Report

**Repository URL**: [https://github.com/ZhuoyueLian/cs6650-hw10-microservices.git](https://github.com/ZhuoyueLian/cs6650-hw10-microservices.git)

## RabbitMQ Queue Length Over Time

### Test Configuration

We conducted a comprehensive load test with the following configuration:

- **Total Requests**: 200,000 checkout operations
- **Concurrent Threads**: 200
- **Warehouse Workers**: 10 worker goroutines per warehouse service instance
- **Test Duration**: 519.38 seconds (~8.7 minutes)

### Load Test Results

The load test completed successfully with the following results:

```
Threads: 200, Total Requests: 200,000
Completed: 200,000 requests
Duration: 519.38 seconds
Successful: 179,814 (89.91%)
Failed: 20,186
Payment Declined: 20,174 (expected ~10%)
Throughput: 346.21 successful requests/second
Error breakdown:
  payment_declined: 20,174
  connection_error: 12
```

**Key Metrics:**
- **Throughput**: 346.21 successful requests/second
- **Success Rate**: 89.91% (10% payment decline is expected behavior)
- **Total Duration**: 519.38 seconds

### Queue Behavior Analysis

The RabbitMQ Management UI dashboard shows the queue behavior over the 10-minute test period:

![RabbitMQ Management UI](/hw10/screenshots/Screenshot%202025-11-16%20at%208.35.11 PM.png)

### Queue Length Pattern

The "Queued messages last ten minutes" graph (top left) reveals:

1. **Peak Queue Length**: The queue reached maximum lengths of approximately 350-420 messages during several spikes (observed around 20:26, 20:27, 20:29, 20:31, and 20:32).

2. **Rapid Processing**: After each peak, the queue length quickly dropped back to near zero, demonstrating efficient message processing by warehouse workers.

3. **Final State**: At the end of the test period (20:34-20:35), the queue length returned to zero, confirming all messages were successfully processed.

4. **Queue Stability**: The maximum queue length of ~420 messages is well below the 1,000 message threshold.

### Message Processing Rates

The "Message rates last ten minutes" graph (bottom left) shows:

1. **Processing Rate**: Message consumption rate fluctuated between 300-400 messages/second during active processing, closely matching the publish rate of ~346 messages/second.

2. **Rate Matching**: The consumption rate matched the expected publish rate, confirming the warehouse service kept up with incoming messages.

3. **Test Completion**: The message rate dropped to near zero at the end (around 20:34), indicating all messages were processed.

### Current Queue Status

At the time of the screenshot, the queue statistics (top right) show:
- **Ready**: 0 messages
- **Unacked**: 0 messages  
- **Total**: 0 messages

This confirms the queue was completely empty with no backlog.

### Key Observations

1. **Dynamic Queue Pattern**: Multiple spikes reaching 350-420 messages, followed by rapid processing to near zero, demonstrating effective handling of traffic bursts.

2. **Near-Zero Final State**: The queue consistently returned to zero after processing cycles, confirming efficient message consumption and no message loss.

3. **Queue Length < 1000**: Maximum queue length of ~420 messages is well below the threshold, confirming adequate processing capacity with 10 warehouse workers.

4. **Efficient Processing**: Rapid drops from peak (420 messages) to near zero demonstrate processing rates matching or exceeding publish rates.

### System Performance Summary

- **Maximum Queue Length**: ~420 messages (peak spikes)
- **Final Queue Length**: 0 messages
- **Throughput**: 346.21 successful requests/second
- **Message Processing Rate**: 300-400 messages/second
- **Queue Pattern**: Dynamic spikes with rapid processing to near-zero
- **System Status**: Healthy - no backlog, no message loss

### Conclusion

The RabbitMQ queue behavior demonstrates effective message processing. The pattern of queue spikes reaching ~420 messages followed by rapid processing to near zero indicates:

1. **Adequate Processing Capacity**: 10 warehouse workers are sufficient to process messages at rates matching or exceeding the publish rate (~346 messages/second).

2. **Effective Buffering**: The queue effectively buffers traffic spikes without processing delays or message loss.

3. **System Stability**: Maximum queue length of 420 messages is well below the 1,000 threshold, confirming healthy operation during peak loads.

4. **Complete Processing**: The final state of zero queued messages confirms all 179,814 successful orders were processed reliably.

The test results confirm the system successfully implements asynchronous message processing with RabbitMQ, achieving a throughput of 346.21 requests/second while maintaining queue stability and near-zero queue length after processing completion.
