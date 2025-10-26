# HW7 Report: Async Order Processing with AWS ECS

## Phase 2: The Bottleneck - Analysis & Documentation

### The Math

**Payment processor speed:** 1 order per 3 seconds = **0.33 orders/second**

**With 20 concurrent customers:**
- Maximum throughput = 20 concurrent requests ÷ 3 seconds = **6.67 orders/second**

**Flash sale demand:** 60 orders/second (from assignment)

**Orders lost per second:** 60 - 6.67 = **53.33 orders/second** ❌

### The Reality

In our Locust testing, the synchronous system showed:
- **Median response time: 3024ms** (customers wait 3 seconds)
- **Actual throughput: ~5.7 orders/second** (far below demand)
- **Result:** Customers experience long waits, many would abandon their carts

### What Happens to Customers?

With synchronous processing, customers face:
1. **Long loading times** (3+ seconds per order)
2. **Browser timeouts** for impatient users
3. **Duplicate orders** from clicking "Submit" multiple times
4. **Lost sales** as customers abandon slow checkout

### The Harsh Reality

**You can't make payment processing faster** - it's fixed at 3 seconds. The only solution: **stop making customers wait**.

---

## Phase 3-5: The Async Solution - Results

### Architecture Change

**Before (Sync):**
```
Customer → API → Payment (3s) → Response ❌ Customer waits
```

**After (Async):**
```
Customer → API → SNS → SQS → Response ✅ Customer gets instant confirmation
                           ↓
                    ECS Workers → Payment (3s)
```

### Performance Comparison

| Metric | Sync | Async | Improvement |
|--------|------|-------|-------------|
| Response Time | 3024ms | 39ms | **77x faster** |
| Success Rate | 100% | 100% | Same |
| User Experience | Wait 3s | Instant | ✅ Better |
| Orders Lost | Many | None | ✅ Perfect |

### Worker Scaling Results

| Workers | Processing Rate | Queue Depth | Can Keep Up? |
|---------|-----------------|-------------|--------------|
| 1 | 0.33/sec | Grows rapidly | ❌ No |
| 5 | 1.67/sec | Still grows | ❌ No |
| 20 | 6.67/sec | Stays at ~0 | ✅ Yes |
| 100 | 33.3/sec | Always 0 | ✅ Yes (overkill) |

### CloudWatch Metrics Analysis

**Key Finding:** With 20+ workers, the queue depth remained at **0 messages** throughout all tests, as shown in the "Approximate Number of Messages Visible" metric.

**What the metrics show:**
- **Messages Received:** 334+ orders successfully queued
- **Messages Deleted:** 337+ orders successfully processed
- **Queue Depth:** Remained at 0 with adequate workers
- **Empty Receives:** Workers polling efficiently with long polling

### Conclusion

**Minimum workers needed:** **20 workers** to handle ~6 orders/second with zero queue buildup.

**For 60 orders/second:** Would need approximately **182 workers** (60 ÷ 0.33 = 182).

**Calculation:**
- Each worker processes 1 order per 3 seconds = 0.33 orders/second
- To handle 60 orders/second: 60 ÷ 0.33 = 182 workers needed

**Key insight:** Async architecture enables 100% order acceptance while background workers process at their own pace, eliminating customer wait times and lost sales.

---

## Analysis Questions

### 1. How many times more orders did your asynchronous approach accept compared to your synchronous approach?

Both approaches accepted the same number of orders (100% success rate), but the async approach accepted them **77x faster** (39ms vs 3024ms response time). The key difference is customer experience: async returns immediately while sync makes customers wait 3 seconds per order.

### 2. What causes queue buildup and how do you prevent it?

**Causes:**
- Incoming order rate exceeds worker processing capacity
- Example: 6 orders/second incoming, but only 0.33 orders/second processing = queue grows by 5.67 orders/second

**Prevention:**
- Scale workers to match or exceed incoming rate
- For our 6 orders/second load: Need at least 18-20 workers (6 ÷ 0.33 = 18)
- For flash sale 60 orders/second: Need ~182 workers

### 3. When would you choose sync vs async in production?

**Choose Sync when:**
- Immediate response required (payment confirmation, authentication)
- Operations are fast (<100ms)
- Simple request-response pattern sufficient

**Choose Async when:**
- Long-running operations (>1 second)
- High traffic spikes expected
- User doesn't need to wait for completion
- Want to decouple services for reliability

**Our case:** Async is clearly better - customers get instant confirmation while payment processing happens in background.

---

## Infrastructure

### AWS Services Deployed

- **VPC:** 10.0.0.0/16 with public/private subnets across 2 AZs
- **Application Load Balancer:** Public-facing endpoint for order API
- **ECS Cluster:** Fargate tasks for order-api and order-processor
- **SNS Topic:** order-processing-events (pub/sub messaging)
- **SQS Queue:** order-processing-queue (message persistence, long polling enabled)
- **ECR:** Container image repositories for both services
- **CloudWatch:** Logs and metrics for monitoring

### Terraform Resources

All infrastructure deployed as code:
- `main.tf` - SNS, SQS, ECR repositories
- `vpc.tf` - VPC, subnets, NAT gateway, routing
- `security_groups.tf` - ALB and ECS task security groups
- `iam.tf` - LabRole configuration
- `alb.tf` - Load balancer and target groups
- `ecs.tf` - ECS cluster, task definitions, services

### Application Architecture

**Order API (order-api):**
- Handles `/orders/sync` and `/orders/async` endpoints
- Publishes orders to SNS topic
- Returns 202 Accepted immediately for async orders

**Order Processor (order-processor):**
- Polls SQS queue with long polling (20s wait time)
- Configurable worker count via `NUM_WORKERS` environment variable
- Processes orders with simulated 3-second payment delay
- Deletes messages after successful processing

---

## Load Testing Results

### Test Configuration
- Tool: Locust
- Users: 20 concurrent
- Spawn rate: 10 users/second
- Duration: 60 seconds per test

### Sync Endpoint Results
- Requests: 321
- Failures: 0
- Median response: 3024ms
- RPS: 5.7

### Async Endpoint Results (100 Workers)
- Requests: 337
- Failures: 0
- Median response: 39ms
- RPS: 5.4
- Queue depth: 0 (workers keeping up perfectly)

---

## Conclusion

Successfully implemented and deployed an event-driven async order processing system that:
- ✅ Accepts 100% of orders with 77x faster response times
- ✅ Scales workers to prevent queue buildup
- ✅ Deployed on AWS ECS with complete IaC using Terraform
- ✅ Monitored with CloudWatch metrics
- ✅ Demonstrates production-ready async architecture patterns