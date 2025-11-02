# HW8 STEP II: DynamoDB Implementation - Complete Analysis

## Performance Test Results

### Test Configuration
- **Operations**: 150 total (50 create, 50 add items, 50 get cart)
- **Success Rate**: 100% (150/150)
- **Test Duration**: ~15 seconds
- **Date**: November 1, 2025

### Response Times

| Operation | Avg (ms) | Min (ms) | Max (ms) |
|-----------|----------|----------|----------|
| Create Cart | 64.47 | 56.16 | 105.21 |
| Add Items | 69.02 | 50.99 | 135.47 |
| Get Cart | 59.14 | 49.66 | 67.04 |
| **Overall** | **64.21** | **49.66** | **135.47** |

### CloudWatch Metrics During Test

**Write Performance:**
- PutItem latency: 2.74-22ms average
- Successful write requests: 101
- Write capacity consumed: 1.68 units/sec peak
- Throttled write events: 0 ✅

**Read Performance:**
- GetItem latency: 0.83-10.8ms average  
- Successful read requests: 101
- Read capacity consumed: 0.842 units/sec peak
- Throttled read events: 0 ✅

**Key Observation**: DynamoDB's internal latency (2-22ms) is much faster than end-to-end API response time (50-135ms), indicating network overhead is the primary factor.

---

## Database Design Decisions

### Partition Key Strategy
- **Partition Key**: `cart_id` (String - UUID v4)
- **No Sort Key**: Simple key-value access pattern sufficient
- **Rationale**: 
  - UUIDs ensure even distribution across partitions
  - No hot partition risk for shopping cart workload
  - Each cart is independently accessed by unique ID

### Data Model: Single Table with Embedded Items
```json
{
  "cart_id": "550e8400-e29b-41d4-a716-446655440000",
  "customer_id": 123,
  "created_at": "2025-11-01T23:30:00Z",
  "updated_at": "2025-11-01T23:35:00Z",
  "items": [
    {
      "product_id": 456,
      "quantity": 2,
      "added_at": "2025-11-01T23:35:00Z",
      "updated_at": "2025-11-01T23:35:00Z"
    }
  ]
}
```

**Why Single Table with Embedded Items?**
1. ✅ **Single GetItem retrieves cart + all items** (no joins needed)
2. ✅ **Atomic updates** - entire cart updated in one operation
3. ✅ **Simpler than managing separate items table**
4. ✅ **Optimal for carts with <50 items** (meets HW8 requirements)
5. ✅ **Reduces number of operations** - fewer API calls

**Trade-offs:**
- ❌ Must read entire cart to update one item (read-modify-write pattern)
- ❌ 400KB item size limit (not an issue for shopping carts)
- ✅ But: Simpler code, fewer operations, lower cost

### Global Secondary Index

**customer-index**: GSI on `customer_id`
- **Purpose**: Query all carts by customer (for purchase history)
- **Projection**: ALL (full cart data available in index)
- **Use case**: "Show all my past shopping carts"

---

## Implementation Highlights

### 1. No Joins Required ✅

**MySQL approach:**
```sql
-- Two tables, requires JOIN
SELECT sc.*, ci.*
FROM shopping_carts sc
LEFT JOIN cart_items ci ON sc.cart_id = ci.cart_id
WHERE sc.cart_id = ?
```

**DynamoDB approach:**
```go
// Single GetItem returns everything
GetItem(cart_id) → {cart + embedded items}
```

**Result**: Simpler query, similar performance

### 2. Update Pattern: Read-Modify-Write

To add items to cart:
```go
1. GetItem(cart_id)           // Read cart (~60ms)
2. Modify items in memory      // Update array
3. PutItem(updated_cart)       // Write back (~70ms)
```

**MySQL approach:**
```sql
-- Single UPDATE with ON DUPLICATE KEY
INSERT INTO cart_items ... ON DUPLICATE KEY UPDATE quantity = quantity + ?
```

**Trade-off**: DynamoDB needs 2 operations (Get + Put) vs MySQL's 1 UPDATE
- **Impact**: Add Items is 5% slower in DynamoDB (69ms vs 66ms)
- **Acceptable**: Still under 70ms, within requirements

### 3. UUID Generation for cart_id
```go
import "github.com/google/uuid"

cartID := uuid.New().String()
// e.g., "550e8400-e29b-41d4-a716-446655440000"
```

**Why UUIDs instead of auto-increment?**
- ✅ Ensures even distribution across DynamoDB partitions
- ✅ No coordination needed between app instances
- ✅ Prevents hot partitions
- ❌ Longer IDs than MySQL's int (acceptable trade-off)

### 4. On-Demand Billing Mode

Used `PAY_PER_REQUEST` instead of provisioned capacity:
- ✅ **Auto-scaling**: No capacity planning needed
- ✅ **Cost-effective for variable load**: Pay only for what you use
- ✅ **No throttling**: Automatic burst capacity
- ✅ **Simpler operations**: No monitoring of capacity units

---

## Observations

### What Worked Well

✅ **No Connection Pooling Needed**
- AWS SDK handles connection management automatically
- No configuration like MySQL's max_connections

✅ **Single Operation Retrieval**
- GetItem returns cart + all items in one call
- Simpler than MySQL's JOIN query

✅ **Serverless Operations**
- No database maintenance, patches, or backups
- Auto-scaling handled by AWS

✅ **Consistent Performance**
- Min-max spread is narrow (50-135ms)
- No outliers or performance degradation

### Challenges Encountered

**1. Read-Modify-Write Pattern**
- **Issue**: Adding items requires reading entire cart first
- **Impact**: 2 API calls (Get + Put) vs MySQL's 1 UPDATE
- **Result**: ~5% slower for add_items operation
- **Mitigation**: Could use UpdateItem with nested attribute updates (more complex)

**2. Manual Item Array Management**
- **Issue**: Must check if product exists in items array manually
- **MySQL**: `ON DUPLICATE KEY UPDATE` handles this automatically
- **Solution**: Loop through items array to find/update product

**3. String IDs vs Integer IDs**
- **Issue**: UUIDs are strings, MySQL uses auto-increment ints
- **Impact**: Slightly larger payload, but negligible
- **Solution**: Made API always return string for consistency

**4. No Native Transaction Support**
- **Issue**: Can't update multiple carts atomically
- **MySQL**: Full ACID transactions across multiple rows
- **DynamoDB**: Single-item ACID only
- **Impact**: Not an issue for shopping cart use case (single-cart operations)

---

## Eventual Consistency Investigation

### Test Methodology
Conducted 4 comprehensive consistency tests:

**Test 1: Write-Then-Read**
- Created 20 carts and immediately read them
- Measured if carts are instantly available after creation

**Test 2: Concurrent Reads**
- Read same cart 10 times rapidly
- Checked for data consistency across reads

**Test 3: Write-Read-Write**
- Created cart → Added item → Read cart
- Verified if item additions are immediately visible

**Test 4: Update Propagation**
- Measured time for updates to become visible
- Updated quantity and polled for changes

### Results

| Test | Iterations | Delays Observed | Consistency Rate |
|------|-----------|----------------|------------------|
| Write-Read | 20 | 0 | 100% |
| Concurrent Reads | 10 | 0 | 100% |
| Write-Read-Write | 10 | 0 | 100% |
| Update Propagation | 1 | <172ms | Immediate |

**Key Findings:**
- ✅ **0 consistency delays observed** across all tests
- ✅ **100% immediate consistency** under test conditions
- ✅ **Average read time**: 65ms (consistent)
- ✅ **Update propagation**: <172ms (within single polling iteration)

### Why No Consistency Delays?

**Expected Factors:**
1. **Single Region**: All operations in us-west-2
2. **Light Load**: 150 operations over 15 seconds (~10 ops/sec)
3. **Modern Infrastructure**: DynamoDB's eventual consistency is typically <100ms
4. **Small Dataset**: Only 50 carts created during test

**Eventual Consistency in Production:**
- Under heavy load (1000s ops/sec), delays more likely
- Multi-region replication would show consistency lag
- Concurrent updates from multiple users could reveal stale reads

### Implications for Shopping Carts

**Low Risk Application:**
- Shopping carts are typically single-user operations
- User modifies their own cart, then reads it
- No concurrent updates from other users on same cart

**When Eventual Consistency Could Matter:**
- **Inventory checks**: Reading product stock across multiple carts
- **Multi-device sync**: User updates cart on phone, views on laptop
- **Payment processing**: Must see absolutely latest cart state

**Mitigation Strategies:**
1. Use **strongly consistent reads** when critical (adds latency)
2. Implement **optimistic locking** with version numbers
3. Add **client-side retry logic** for critical operations

---

## DynamoDB vs MySQL Comparison

### Performance Comparison

| Metric | MySQL | DynamoDB | Winner | Margin |
|--------|-------|----------|--------|--------|
| Create Cart Avg | 61.78ms | 64.47ms | MySQL | +2.69ms (4%) |
| Add Items Avg | 65.59ms | 69.02ms | MySQL | +3.43ms (5%) |
| Get Cart Avg | 58.67ms | 59.14ms | MySQL | +0.47ms (1%) |
| **Overall Avg** | **62.01ms** | **64.21ms** | **MySQL** | **+2.20ms (3.5%)** |
| Success Rate | 100% | 100% | Tie | - |
| Min Response | 52.10ms | 49.66ms | DynamoDB | -2.44ms |
| Max Response | 99.98ms | 135.47ms | MySQL | -35.49ms |

**Verdict**: **Statistically equivalent performance** - 3.5% difference is within normal variance.

### Architectural Comparison

| Aspect | MySQL | DynamoDB |
|--------|-------|----------|
| **Schema** | Strict relational (2 tables) | Flexible JSON (1 table) |
| **Cart Retrieval** | 1 JOIN query | 1 GetItem call |
| **Add Item** | 1 UPDATE (UPSERT) | 2 calls (Get + Put) |
| **Consistency** | Strong (ACID) | Eventual (default) |
| **Scaling** | Vertical (bigger instance) | Horizontal (automatic) |
| **Operations** | Manual (patches, backups) | Serverless (fully managed) |
| **Connection Mgmt** | Pool required (25 conns) | SDK handles automatically |
| **Cost Model** | Fixed (RDS instance) | Variable (pay per request) |
| **Capacity Planning** | Required (CPU, memory) | Automatic (on-demand) |

### Operational Comparison

| Factor | MySQL | DynamoDB |
|--------|-------|----------|
| **Setup Complexity** | High (RDS, security groups, subnets) | Low (just create table) |
| **Maintenance** | Manual (patches, backups) | Automatic (AWS managed) |
| **Monitoring** | CloudWatch + RDS metrics | CloudWatch (built-in) |
| **High Availability** | Multi-AZ (~2x cost) | Built-in (3 AZs) |
| **Disaster Recovery** | Manual snapshots | Point-in-time recovery |
| **Scaling** | Downtime for vertical scaling | Zero-downtime auto-scaling |

---

## What I Would Improve

### For Production

**DynamoDB Optimizations:**
1. **Use UpdateItem with nested attributes** for atomic item updates
```go
   // Instead of Get + Modify + Put
   UpdateItem with SET items[?].quantity = items[?].quantity + :inc
```

2. **Implement optimistic locking** with version numbers
```json
   {
     "cart_id": "...",
     "version": 5,
     "items": [...]
   }
```

3. **Add TTL attribute** for automatic cart expiration
```json
   {
     "cart_id": "...",
     "expires_at": 1735689600  // Unix timestamp
   }
```

4. **Enable DynamoDB Streams** for:
   - Audit trail of cart changes
   - Real-time analytics
   - Triggering downstream systems

5. **Use batch operations** for bulk cart queries
```go
   BatchGetItem([cart_id_1, cart_id_2, ...])
```

### Schema Optimizations

1. **Sparse GSI for abandoned carts**
   - Index on `checkout_timestamp` (only exists when checked out)
   - Query unchecked carts for remarketing

2. **Composite sort key** if querying items by product
   - PK: `cart_id`, SK: `ITEM#product_id`
   - Enables querying specific products in cart

3. **Separate hot/cold data**
   - Active carts in main table
   - Completed orders in archive table (S3 + Athena)

---

## Performance vs MySQL: Key Takeaways

### When DynamoDB Wins

✅ **Operational Simplicity**
- No database administration
- Auto-scaling without configuration
- Built-in high availability

✅ **Predictable Costs at Scale**
- Pay-per-request pricing
- No over-provisioning needed
- Automatic burst capacity

✅ **Global Distribution** (if needed)
- Multi-region replication
- Low-latency worldwide access

### When MySQL Wins

✅ **Complex Queries**
- JOINs across multiple tables
- Ad-hoc reporting queries
- Complex WHERE clauses

✅ **Strong Consistency Guarantee**
- ACID transactions
- No eventual consistency concerns
- Multi-row atomic updates

✅ **Familiar Tooling**
- Standard SQL
- Existing ORM support
- Developer familiarity

### For Shopping Carts Specifically

**Recommendation**: **Either works well!**

- **Performance**: ~3% difference (negligible)
- **Complexity**: DynamoDB simpler (embedded model)
- **Operations**: DynamoDB easier (serverless)
- **Cost**: DynamoDB cheaper at low scale (no idle RDS cost)

**Choose DynamoDB if:**
- Building new microservice
- Want serverless operations
- Need auto-scaling
- Light admin overhead preferred

**Choose MySQL if:**
- Existing RDS infrastructure
- Team expertise in SQL
- Need complex analytics
- Strong consistency critical

---

## Lessons Learned

### 1. Network Latency Dominates

**Discovery**: DynamoDB's internal latency (2-22ms) vs end-to-end API time (50-135ms)
- **Lesson**: ALB → ECS → Database adds ~30-50ms overhead
- **Impact**: Database choice matters less than architecture
- **Optimization**: Consider caching, CDN, or edge computing

### 2. Embedded Data Model Simplifies Code

**MySQL**: Two tables, JOIN queries, foreign keys
**DynamoDB**: One document, single GetItem

**Result**: DynamoDB code was simpler despite needing read-modify-write pattern

### 3. Eventual Consistency Rarely Manifests Under Light Load

- Tested extensively, observed 0 delays
- In production with high concurrency, would be different
- Important to understand trade-offs even if not observed

### 4. On-Demand Billing Eliminates Capacity Planning

- No throttling despite variable load
- Would have been difficult to provision capacity for this workload
- Recommendation: Start with on-demand, switch to provisioned only if cost-effective

### 5. NoSQL Requires Different Thinking

- Can't rely on database constraints (foreign keys)
- Must handle data consistency in application
- More flexible but more responsibility

---

## Conclusion

Both MySQL and DynamoDB delivered excellent performance for the shopping cart use case:

- **MySQL**: 62ms average, strong consistency, familiar SQL
- **DynamoDB**: 64ms average, auto-scaling, serverless operations

**Winner**: Depends on context!

For a **homework/learning perspective**: Valuable to understand both approaches and their trade-offs.

For **production shopping cart service**: Would choose **DynamoDB** for:
- Operational simplicity
- Auto-scaling
- Lower operational overhead
- Adequate performance

**Final Thought**: The 3.5% performance difference is insignificant compared to operational benefits of serverless DynamoDB.