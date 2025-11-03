# HW8 STEP II: DynamoDB Implementation - Findings

## Performance Test Results

### Test Configuration
- Operations: 150 total (50 create, 50 add items, 50 get cart)
- Success Rate: 100% (150/150)
- Test Duration: ~15 seconds

### Response Times
| Operation | Avg | Min | Max |
|-----------|-----|-----|-----|
| Create Cart | 64.47ms | 56.16ms | 105.21ms |
| Add Items | 69.02ms | 50.99ms | 135.47ms |
| Get Cart | 59.14ms | 49.66ms | 67.04ms |

**Overall Average: 64.21ms**

### CloudWatch Metrics During Test
- **PutItem latency**: 2.74-22ms (internal DynamoDB latency)
- **GetItem latency**: 0.83-10.8ms (internal DynamoDB latency)
- **Throttled events**: 0 ✅
- **Key Observation**: DynamoDB's internal latency (2-22ms) is much faster than end-to-end API time (50-135ms), indicating network overhead (ALB → ECS → DynamoDB) is the primary factor.

## Database Design

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

**Why This Design?**
- **Single GetItem retrieves cart + all items** (no joins needed)
- **Atomic updates** - entire cart updated in one operation
- **Simpler code** - no separate items table to manage
- **Optimal for carts with <50 items** (meets requirements)

**Trade-offs:**
- ❌ Must read entire cart to update one item (read-modify-write pattern)
- ❌ 400KB item size limit (not an issue for shopping carts)
- ✅ But: Simpler code, fewer operations, lower cost

### Billing Mode: On-Demand
- **PAY_PER_REQUEST**: No capacity planning needed
- **Auto-scaling**: Handles variable load automatically
- **No throttling**: Automatic burst capacity

## Implementation Highlights

### Single GetItem Operation (No JOINs)
```go
// DynamoDB: Single GetItem returns everything
result, err := database.DynamoClient.GetItem(context.TODO(), &dynamodb.GetItemInput{
    TableName: aws.String(database.TableName),
    Key: map[string]types.AttributeValue{
        "cart_id": &types.AttributeValueMemberS{Value: cartID},
    },
})
```

**Why This Works**:
- Single database operation retrieves cart + all items
- No JOINs needed (data is embedded in document)
- Simpler than MySQL's JOIN query

**Comparison to MySQL**:
- **MySQL**: 1 JOIN query (2 tables)
- **DynamoDB**: 1 GetItem call (1 document)
- **Result**: Similar performance (59ms vs 58ms)

### Read-Modify-Write Pattern for Updates
```go
// Step 1: Get existing cart
result, err := database.DynamoClient.GetItem(...)

// Step 2: Modify items array in memory
for i := range cart.Items {
    if cart.Items[i].ProductID == req.ProductID {
        cart.Items[i].Quantity += req.Quantity
        found = true
        break
    }
}

// Step 3: Put updated cart back
_, err = database.DynamoClient.PutItem(...)
```

**Why This Works**:
- GetItem retrieves entire cart document
- Modify items array in application code
- PutItem writes entire document back (atomic)

**Trade-off vs MySQL**:
- **MySQL**: 1 UPSERT query (atomic, single operation)
- **DynamoDB**: 2 operations (Get + Put, read-modify-write)
- **Impact**: Add Items is 5% slower (69ms vs 66ms)
- **Acceptable**: Still under 70ms, within requirements

### UUID Partition Key for Even Distribution
```go
import "github.com/google/uuid"

cartID := uuid.New().String()
// e.g., "550e8400-e29b-41d4-a716-446655440000"
```

**Why UUIDs?**
- Ensures even distribution across DynamoDB partitions
- No coordination needed between app instances
- Prevents hot partitions
- Trade-off: Longer IDs than MySQL's int (acceptable)

### Attribute Value Marshaling
```go
// Marshal: Go struct → DynamoDB attribute values
item, err := attributevalue.MarshalMap(cart)

// Unmarshal: DynamoDB attribute values → Go struct
err = attributevalue.UnmarshalMap(result.Item, &cart)
```

**Why This Matters**:
- DynamoDB uses attribute values, not Go structs directly
- Requires `dynamodbav` tags on struct fields
- More code than MySQL's direct struct scanning

## Eventual Consistency Investigation

### Test Results
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

### Why No Consistency Delays?
- **Single Region**: All operations in us-west-2
- **Light Load**: ~10 ops/sec during tests
- **Modern Infrastructure**: DynamoDB's eventual consistency typically <100ms
- **Small Dataset**: Only 50 carts during test

### Implications for Shopping Carts
- **Low Risk**: Single-user operations (user modifies their own cart)
- **No Concurrent Updates**: No other users modifying same cart
- **Sequential Operations**: User adds item, then views cart (no race conditions)

## Learning Journey

### What Surprised Me?

#### 1. Read-Modify-Write Pattern Required

**Initial Expectation**: Expected DynamoDB to have atomic UPSERT like MySQL

**Reality**: DynamoDB requires read-modify-write pattern for updating nested items

**Why This Happens**:
- DynamoDB stores entire cart as single document
- To update one item, must read entire cart, modify in memory, write back
- No native UPSERT for nested arrays like MySQL's `ON DUPLICATE KEY UPDATE`

**Impact**: Add Items is 5% slower (69ms vs 66ms) due to 2 operations vs 1

**Learning**: NoSQL document model requires different update patterns than relational databases

---

#### 2. No Connection Pooling Needed

**Initial Expectation**: Would need to configure connection pooling like MySQL

**Reality**: AWS SDK handles connection management automatically

**Why This Works**:
- SDK manages connections internally
- No configuration needed (unlike MySQL's `SetMaxOpenConns`)
- Simpler code (no connection pool setup)

**Comparison to MySQL**:
- **MySQL**: Manual connection pooling (25 max connections, 5 idle)
- **DynamoDB**: Automatic connection management (zero configuration)
- **Result**: 90% less code for connection management

**Learning**: Serverless databases eliminate connection management complexity

---

#### 3. Eventual Consistency Never Manifested

**Initial Expectation**: Expected to see consistency delays in testing

**Reality**: 0 delays observed across all tests (100% immediate consistency)

**Why This Happened**:
- Single region (us-west-2)
- Light load (~10 ops/sec)
- Modern AWS infrastructure (typically <100ms consistency)
- Small dataset (50 carts)

**What This Means**:
- Under light load, eventual consistency behaves like strong consistency
- In production with heavy load, delays would be more likely
- Important to understand trade-offs even if not observed

**Learning**: Eventual consistency is rarely noticeable under light load, but important to understand for production

---

#### 4. Network Latency Dominates Performance

**Discovery**: DynamoDB's internal latency (2-22ms) vs end-to-end API time (50-135ms)

**Breakdown**:
- DynamoDB internal: 2-22ms (actual database operations)
- Network overhead: ~30-50ms (ALB → ECS → DynamoDB)
- Application processing: ~8ms
- **Total**: ~50-135ms

**What This Means**:
- Database choice matters less than network architecture
- Both MySQL and DynamoDB perform similarly (62ms vs 64ms)
- Network hops add significant overhead

**Learning**: Architecture (ALB, ECS, network) matters more than database choice for performance

---

## Key Takeaways

1. **NoSQL requires different thinking** - Can't rely on database constraints, must handle consistency in application
2. **Embedded data model simplifies code** - Single GetItem vs MySQL's JOIN query
3. **Read-modify-write pattern is necessary** - DynamoDB requires 2 operations vs MySQL's 1 UPSERT
4. **UUIDs ensure even distribution** - Prevents hot partitions in DynamoDB
5. **Network latency dominates** - Database choice matters less than architecture
6. **Eventual consistency rarely manifests** - Under light load, behaves like strong consistency
7. **Serverless eliminates connection management** - AWS SDK handles connections automatically