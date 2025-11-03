# Step 3: Database Comparison & Analysis - Complete Report

## Executive Summary

This report presents a comprehensive comparison between MySQL (RDS) and DynamoDB for a shopping cart service. Through hands-on implementation and testing, I evaluated performance, resource efficiency, consistency models, and operational complexity. **Key Finding:** Both databases perform almost identically (62ms vs 64ms), but operational overhead differs significantly. DynamoDB offers 90% simpler code and zero connection management, while MySQL provides fixed costs and SQL query flexibility.

**Recommendation:** Choose DynamoDB for operational simplicity and auto-scaling, choose MySQL for fixed budgets and complex queries.

---

## Part 0: Data Verification and Merging

### Objective
Verify both test datasets contain exactly 150 operations and create a unified dataset for analysis.

### Methodology
- Loaded `mysql_test_results.json` and `dynamodb_test_results.json`
- Verified operation counts and distribution by operation type
- Merged datasets into `combined_results.json` with database identifiers
- Generated `data_verification.json` for verification metadata

### Results

**Data Verification:**
- **MySQL:** 150 operations total (50 create_cart, 50 add_items, 50 get_cart)
- **DynamoDB:** 150 operations total (50 create_cart, 50 add_items, 50 get_cart)
- **Status:** ✅ Both datasets valid and consistent

**Merged Dataset:**
- **Total Operations:** 300 (150 MySQL + 150 DynamoDB)
- **Format:** Flat array with `database` field identifier
- **Output:** `combined_results.json` (single source for all analysis)

---

## Part 1: Performance Comparison Analysis

### Test Configuration
- **Total Operations:** 300 (150 per database)
- **Operations:** 50 create_cart, 50 add_items, 50 get_cart per database
- **Test Duration:** ~15 seconds per database
- **Success Rate:** 100% for both databases

### Performance Results

| Metric | MySQL | DynamoDB | Winner | Difference |
|--------|-------|----------|--------|------------|
| **Average Response Time** | 62.01ms | 64.21ms | MySQL | +2.20ms (3.5%) |
| **Min Response Time** | 52.10ms | 49.66ms | DynamoDB | -2.44ms |
| **Max Response Time** | 99.98ms | 135.47ms | MySQL | -35.49ms |
| **P50 (Median)** | 60.90ms | 62.66ms | MySQL | +1.76ms |
| **P95** | 71.21ms | 73.85ms | MySQL | +2.64ms |
| **P99** | 96.67ms | 114.86ms | MySQL | +18.19ms |
| **Success Rate** | 100% | 100% | Tie | - |

### Operation-Specific Breakdown

| Operation | MySQL Avg | DynamoDB Avg | Difference | Winner |
|-----------|-----------|--------------|------------|--------|
| **Create Cart** | 61.78ms | 64.47ms | +2.69ms (4%) | MySQL |
| **Add Items** | 65.59ms | 69.02ms | +3.43ms (5%) | MySQL |
| **Get Cart** | 58.67ms | 59.14ms | +0.47ms (1%) | MySQL |

### Key Findings

1. **Performance Difference is Negligible:** 2.20ms average difference (3.5%) is within normal variance
2. **Both Achieved 100% Success Rate:** No operational failures observed
3. **MySQL Slightly Faster:** Better P99 (96.67ms vs 114.86ms) indicates more consistent tail latency
4. **DynamoDB Had Wider Variance:** Max response time 135.47ms vs MySQL's 99.98ms

**Verdict:** Performance is statistically equivalent. The choice between databases should focus on operational complexity, scaling strategy, and cost model rather than raw performance.

### Consistency Model Impact Assessment

**Investigation Results:**

1. **Actual Consistency Behavior Observed:**
   - **MySQL:** 100% strong consistency (ACID guarantees)
   - **DynamoDB:** 0 consistency delays observed (100% immediate consistency under test load)
   - **Evidence:** All write-then-read operations returned consistent data immediately

2. **DynamoDB Eventual Consistency Impact:**
   - **No Practical Impact:** Observed zero delays across all tests
   - **Reason:** Light load (~10 ops/sec) + single region = consistency delays unlikely
   - **Shopping Cart Use Case:** Single-user operations naturally avoid consistency issues

3. **Consistency Guarantee Comparison:**

| Aspect | MySQL (ACID) | DynamoDB (Eventual) |
|--------|-------------|---------------------|
| **Guarantees** | Strong consistency, transactions | Eventual consistency (typically <100ms) |
| **Updates** | Immediately visible to all readers | May take <100ms to propagate |
| **Transactions** | Multi-row ACID transactions | Single-item ACID only |
| **Complexity** | Built-in constraints, foreign keys | Application handles consistency |

4. **User Experience Implications:**
   - **Low Risk for Shopping Carts:** Users modify their own cart, then read it
   - **Eventual Consistency Rarely Noticed:** Single-user scenarios avoid consistency issues
   - **Higher Risk Scenarios:** Multi-device sync, inventory checks across carts, payment processing

**Key Insight:** For shopping carts specifically, eventual consistency had zero observable impact due to single-user access patterns and light load. The pattern (user modifies own cart, then reads it) naturally avoids consistency issues.

---

## Part 2: Resource Efficiency Analysis

### Connection Management Overhead

**MySQL:**
- **Configuration:** MaxOpenConns: 25, MaxIdleConns: 5, ConnMaxLifetime: 5 min
- **Code Complexity:** ~40 lines (connection pooling, retry logic, schema initialization)
- **Observed Usage:** Peak 6 connections used (configured for 25) - pool had 19 unused connections
- **Overhead:** Connection establishment ~50ms without pooling, ~5ms with pooling
- **Scaling:** Must adjust pool size manually as load increases

**DynamoDB:**
- **Configuration:** Zero - AWS SDK handles automatically
- **Code Complexity:** 2 lines (SDK initialization: `config.LoadDefaultConfig()` and `dynamodb.NewFromConfig()`)
- **Observed Usage:** No connection management concerns, zero throttling events
- **Overhead:** Zero - connections scale automatically with request volume
- **Scaling:** Automatic - no manual intervention needed

**Winner:** DynamoDB - Zero operational overhead (90% simpler code)

### Resource Predictability and Capacity Planning

**MySQL (RDS):**
- **Cost Model:** Fixed ~$15/month for db.t3.micro, regardless of usage
- **Resource Usage:** Predictable CPU/memory patterns (3.5% CPU during tests, 110-115MB freeable memory)
- **Capacity Planning:** Must provision for peak load upfront (e.g., Black Friday traffic)
- **Scaling:** Vertical scaling (bigger instance = requires downtime)
- **Observation:** RDS CPU at 3.5% shows lots of headroom - paying for unused capacity

**DynamoDB (On-Demand):**
- **Cost Model:** Pay-per-request (~$0.25 per million write requests, ~$0.25 per million read requests)
- **Resource Usage:** No fixed resources to monitor
- **Capacity Planning:** None required - auto-scales to handle any load instantly
- **Scaling:** Horizontal scaling (zero-downtime, automatic)
- **Observation:** DynamoDB internal latency (2-22ms) stayed consistent, no throttling

**Winner:** MySQL for cost predictability, DynamoDB for performance predictability

### Operational Complexity Differences

**MySQL Setup:**
- **Terraform:** RDS module, security groups, subnet groups (4+ resources)
- **Go Code:** Connection pooling, retry logic, schema initialization (~100 lines)
- **Monitoring:** CloudWatch RDS metrics (CPU, connections, memory)
- **Maintenance:** Schema migrations, backups, patches

**DynamoDB Setup:**
- **Terraform:** Single table resource (1 resource)
- **Go Code:** AWS SDK initialization (~10 lines)
- **Monitoring:** CloudWatch DynamoDB metrics (latency, throttling) - optional
- **Maintenance:** None (AWS managed)

**Winner:** DynamoDB - Significantly simpler setup and maintenance

### Scaling Analysis

**How Resource Requirements Change with Load:**

| Load Level | MySQL (RDS) | DynamoDB |
|------------|-------------|----------|
| **Low (10 ops/sec)** | 3.5% CPU, 6 connections, fixed $15/mo | Minimal cost (~$0.01/month), auto-scales |
| **Medium (100 ops/sec)** | ~20-30% CPU, 15-20 connections, same $15/mo | Higher cost (~$0.10/month), still auto-scales |
| **High (1000 ops/sec)** | Hits CPU limits → needs bigger instance | Auto-scales instantly (~$1/month), pay-per-request |
| **Scaling Method** | Manual vertical scaling (downtime) | Automatic horizontal (zero downtime) |

**Which Offers More Predictable Resource Consumption?**

- **Cost Predictability:** MySQL wins (fixed $15/month)
- **Performance Predictability:** DynamoDB wins (consistent latency regardless of load)
- **Operational Predictability:** DynamoDB wins (no manual intervention)

**Capacity Planning Implications:**

- **MySQL:** Requires proactive capacity planning, risk of over/under-provisioning
- **DynamoDB:** No capacity planning needed, reactive cost optimization (if needed)

**Key Insight:** MySQL requires proactive capacity planning and manual scaling, while DynamoDB automatically handles load spikes without configuration.

---

## Part 3: Real-World Scenario Recommendations

### Scenario A: Startup MVP
**Context:** 100 users/day, 1 developer, limited budget, quick launch

**Recommendation:** **DynamoDB**

**Key Evidence:**
- **Cost:** DynamoDB ~$0.01/month vs MySQL $15/month fixed - **1500x cheaper** at this scale
- **Setup:** DynamoDB takes 5 minutes vs MySQL needs RDS setup, security groups, subnet groups
- **Operations:** DynamoDB requires zero maintenance (1 developer can focus on features)
- **Performance:** Both perform similarly (64.21ms vs 62.01ms - only 3.5% difference)

**Why Not MySQL:** Fixed $15/month cost for very low usage, more setup complexity, developer time spent on database ops

---

### Scenario B: Growing Business
**Context:** 10K users/day, 5 developers, moderate budget, feature expansion

**Recommendation:** **MySQL**

**Key Evidence:**
- **Cost Predictability:** MySQL fixed $15/month vs DynamoDB variable cost (~$10-15/month at this scale)
- **Team Expertise:** 5 developers likely familiar with SQL (easier onboarding)
- **Complex Queries:** Growing business needs analytics, reporting (MySQL JOINs better suited)
- **Performance:** MySQL showed better P99 (96.67ms vs 114.86ms) and more consistent performance

**Why Not DynamoDB:** Variable cost can surprise as traffic grows, complex queries require application-level joins, team might need DynamoDB training

---

### Scenario C: High-Traffic Events
**Context:** 50K normal, 1M spike users, revenue-critical, can invest in infrastructure

**Recommendation:** **DynamoDB**

**Key Evidence:**
- **Auto-Scaling:** DynamoDB scales automatically to 1M spike without configuration (zero downtime)
- **Performance Consistency:** DynamoDB maintained 59-69ms even with variable load (no degradation)
- **Cost Efficiency:** DynamoDB pay-per-request means only pay for spike when it happens (~$250 for 1M spike vs $300/month for idle MySQL)
- **Evidence:** Zero throttling events observed, auto-burst capacity handled variable load seamlessly

**Why Not MySQL:** Must provision for peak (1M spike) = expensive idle capacity, vertical scaling requires downtime (revenue-critical = unacceptable)

---

### Scenario D: Global Platform
**Context:** Millions of users, multi-region, 24/7 availability, enterprise requirements

**Recommendation:** **DynamoDB**

**Key Evidence:**
- **Multi-Region:** DynamoDB global tables provide automatic multi-region replication (MySQL requires complex read replicas setup)
- **24/7 Availability:** DynamoDB built-in HA across 3 AZs, zero maintenance downtime
- **Enterprise Features:** Point-in-time recovery, automatic backups, encryption (all built-in)
- **Scaling:** Handles millions of users without manual intervention (auto-scales horizontally)

**Why Not MySQL:** Multi-region MySQL requires read replicas, complex replication setup, vertical scaling downtime unacceptable for 24/7 enterprise service

### Summary of Recommendations

| Scenario | Recommendation | Primary Reason |
|----------|---------------|---------------|
| A: Startup MVP | DynamoDB | Cost + simplicity |
| B: Growing Business | MySQL | Cost predictability + SQL expertise |
| C: High-Traffic Events | DynamoDB | Auto-scaling + zero downtime |
| D: Global Platform | DynamoDB | Multi-region + operational simplicity |

**Pattern:** Choose DynamoDB when operational simplicity, auto-scaling, or global distribution matters. Choose MySQL when fixed budget, team SQL expertise, or complex queries needed.

---

## Part 4: Evidence-Based Architecture Recommendations

### Shopping Cart Winner

**Recommendation:** **DynamoDB**

**Why:**
- Operational simplicity (zero connection management overhead)
- Auto-scaling (handles traffic spikes without downtime)
- Cost-effective at low scale (1500x cheaper than MySQL for startups)

**Supporting Evidence:**

1. **Performance:** DynamoDB 64.21ms avg vs MySQL 62.01ms - **Response time advantage: MySQL faster by 2.20ms (3.5% difference - negligible)**

2. **Implementation Complexity:**
   - MySQL: ~100 lines (connection pooling, retry logic, schema initialization)
   - DynamoDB: ~10 lines (simple SDK initialization)
   - **Complexity difference: DynamoDB 90% simpler code**

3. **Operational Overhead:**
   - MySQL: Requires connection pool monitoring, schema migrations
   - DynamoDB: Zero maintenance, AWS manages everything
   - **Evidence:** DynamoDB handled all requests without throttling, zero configuration

4. **Other Factors:**
   - **Cost:** DynamoDB ~$0.01/month at low scale vs MySQL $15/month fixed
   - **Scaling:** DynamoDB auto-scales, MySQL requires manual vertical scaling (downtime)
   - **Success Rate:** Both 100% (tie)

**Verdict:** For shopping carts specifically, DynamoDB's operational simplicity outweighs the 2.20ms performance difference.

### When to Choose MySQL Instead

**Despite recommending DynamoDB, choose MySQL when:**

1. **Fixed Budget Required:** Need predictable monthly costs
   - Evidence: MySQL fixed $15/month vs DynamoDB variable cost
   - Use Case: Growing business with budget constraints

2. **Complex Queries Needed:** Need JOINs across multiple tables, analytics, reporting
   - Evidence: MySQL SQL queries easier than DynamoDB's read-modify-write pattern
   - Use Case: Business intelligence, customer analytics

3. **Team SQL Expertise:** Team familiar with SQL, not NoSQL
   - Evidence: MySQL requires standard SQL knowledge (easier onboarding)
   - Use Case: Team without DynamoDB training budget

4. **Multi-Row Transactions:** Need ACID transactions across multiple rows/tables
   - Evidence: MySQL full ACID support vs DynamoDB single-item only
   - Use Case: Order processing, inventory management

### Polyglot Strategy: Using Both Databases

If building a complete e-commerce system, here's how I'd use both:

**Shopping Carts: DynamoDB**
- Simple access patterns (create, read, update by cart_id)
- Auto-scaling for traffic spikes (Black Friday)
- Low operational overhead (zero maintenance)
- **Evidence:** Tests showed 100% success rate, zero throttling, auto-scales seamlessly

**User Sessions: DynamoDB**
- Similar to shopping carts (simple key-value access)
- High write volume (frequent session updates)
- Natural expiration (TTL support)
- **Pattern:** Same partition key strategy (session_id), auto-expires

**Product Catalog: MySQL**
- Complex queries (search by category, price range, filters)
- JOINs needed (products + categories + reviews)
- Relatively stable data (less frequent updates)
- **Pattern:** Requires SQL queries, joins, full-text search

**Order History: MySQL**
- Complex reporting queries (sales by date, customer history)
- JOINs needed (orders + order_items + customers)
- ACID transactions critical (order consistency)
- **Pattern:** Requires SQL analytics, transactions across tables

**Summary:** Use DynamoDB for high-write, simple-access patterns (carts, sessions). Use MySQL for complex queries, analytics, and transactions (catalog, orders).

---

## Part 5: Learning Reflection

### What Surprised Me?

**1. DynamoDB Performed Almost Identically to MySQL**
- **Expectation:** Thought DynamoDB would be much faster (NoSQL = fast, right?)
- **Reality:** DynamoDB 64.21ms vs MySQL 62.01ms - only 2.20ms difference (3.5%)
- **Surprise:** Performance difference was negligible - operational overhead mattered more than raw speed
- **Evidence:** Both achieved 100% success rate, similar response times across all operations

**2. Connection Pooling Overhead Was Significant**
- **Expectation:** Thought connection pooling would be simple to set up
- **Reality:** Required ~40 lines of code (MaxOpenConns, MaxIdleConns, retry logic)
- **Surprise:** DynamoDB had zero connection management overhead - just 2 lines (SDK initialization)
- **Evidence:** MySQL needed configuration for pool size (25 max, 5 idle), DynamoDB handled it automatically

**3. DynamoDB Schema Design Was Simpler Than Expected**
- **Expectation:** Thought NoSQL schema design would be complex
- **Reality:** DynamoDB schema was simpler - just partition key (cart_id)
- **Surprise:** MySQL required normalized tables (shopping_carts + cart_items) with foreign keys
- **Evidence:** DynamoDB single table vs MySQL two tables with JOIN queries

**4. MySQL Was Actually Faster (Slightly)**
- **Expectation:** Thought DynamoDB would be faster due to NoSQL design
- **Reality:** MySQL averaged 62ms vs DynamoDB 64ms - MySQL was faster!
- **Surprise:** ACID transactions and optimized JOINs performed better than expected
- **Evidence:** Performance analysis showed MySQL 2.20ms faster across all operations

### Key Insights Gained

**When Would I Definitely Choose MySQL?**
1. **Fixed Budget Required:** Need predictable $15/month cost vs variable DynamoDB pricing
2. **Complex Queries Needed:** Need JOINs, analytics, reporting (SQL is better suited)
3. **Team SQL Expertise:** Team familiar with SQL, not NoSQL (easier onboarding)
4. **Multi-Row Transactions:** Need ACID transactions across multiple rows/tables
5. **Evidence:** MySQL showed better P99 (96.67ms vs 114.86ms) and more consistent performance

**When Would I Definitely Choose DynamoDB?**
1. **Operational Simplicity:** Want zero connection management, zero maintenance overhead
2. **Auto-Scaling Needed:** Traffic spikes unpredictable (Black Friday, viral events)
3. **Global Distribution:** Need multi-region replication (global tables)
4. **Startup MVP:** Limited budget, 1 developer, need quick launch
5. **Evidence:** DynamoDB handled all requests without throttling, zero configuration, 90% simpler code

**What Would I Tell Another Student Starting This Assignment?**

**Advice:**
1. **Start with MySQL First:** It's easier to understand (SQL is familiar), then move to DynamoDB
2. **Test Each Database Separately:** Run performance tests for MySQL first, then DynamoDB
3. **Connection Pooling Matters:** For MySQL, connection pooling is critical - don't skip it
4. **Document Everything:** Keep notes on errors you encounter - they're valuable learning

**Key Tip:** The performance difference (2ms) is negligible - focus on operational overhead, scaling, and cost models instead.

**How Did Hands-On Implementation Change My Understanding?**

**Before Implementation:**
- Thought DynamoDB would be much faster (NoSQL = fast)
- Thought connection pooling was simple
- Thought schema design would be similar
- Thought performance would be the deciding factor

**After Implementation:**
- **Performance is similar:** Both perform almost identically (64ms vs 62ms)
- **Operations matter more:** DynamoDB's zero operational overhead was the real win
- **Schema design differs:** DynamoDB simpler (single table) vs MySQL normalized (JOINs)
- **Scaling matters:** DynamoDB auto-scales seamlessly, MySQL requires manual intervention

**Key Realization:** The choice between MySQL and DynamoDB isn't about performance - it's about operational complexity, scaling strategy, and cost model. Performance is nearly identical, but operational overhead differs significantly.

---

## Conclusion

### Key Findings Summary

1. **Performance:** Both databases perform almost identically (62ms vs 64ms average) - difference is negligible (3.5%)

2. **Operational Overhead:** DynamoDB requires 90% less code (10 lines vs 100 lines) and zero connection management

3. **Scaling:** DynamoDB auto-scales seamlessly (zero downtime), MySQL requires manual vertical scaling (downtime)

4. **Cost:** At low scale, DynamoDB 1500x cheaper (~$0.01/month vs $15/month). At high scale, MySQL more predictable.

5. **Consistency:** MySQL provides ACID guarantees. DynamoDB eventual consistency had zero observable impact for shopping carts (single-user access patterns).

### Final Recommendation

**For Shopping Carts:** Choose DynamoDB for operational simplicity, auto-scaling, and cost-effectiveness at startup scale. Choose MySQL for fixed budgets, SQL team expertise, and complex query requirements.

**For Complete E-Commerce Systems:** Use a polyglot approach - DynamoDB for high-write, simple-access patterns (carts, sessions), MySQL for complex queries and analytics (catalog, orders).

### Main Takeaway

**The choice between MySQL and DynamoDB isn't about performance - it's about operational complexity, scaling strategy, and cost model.** Both databases perform similarly (2ms difference is negligible), but operational overhead differs significantly. Focus on your team's needs: operational simplicity vs cost predictability, auto-scaling vs SQL expertise.
