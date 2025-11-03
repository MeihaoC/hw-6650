"""
HW8 Step 3 Part 0: Create Combined Results
Merges mysql_test_results.json and dynamodb_test_results.json into combined_results.json
Verifies both datasets have exactly 150 operations
"""

import json
from collections import defaultdict

# Read both test files
print("📂 Loading test results...")
try:
    with open('mysql_test_results.json', 'r') as f:
        mysql_results = json.load(f)
    with open('dynamodb_test_results.json', 'r') as f:
        dynamodb_results = json.load(f)
    print("✅ Files loaded successfully")
except FileNotFoundError as e:
    print(f"❌ Error: {e}")
    exit(1)
except json.JSONDecodeError as e:
    print(f"❌ Error parsing JSON: {e}")
    exit(1)

# Verify data consistency
print("\n" + "=" * 60)
print("Data Verification")
print("=" * 60)

mysql_count = len(mysql_results)
dynamodb_count = len(dynamodb_results)

# Count by operation type
mysql_ops = defaultdict(int)
dynamodb_ops = defaultdict(int)

for op in mysql_results:
    mysql_ops[op.get('operation', 'unknown')] += 1

for op in dynamodb_results:
    dynamodb_ops[op.get('operation', 'unknown')] += 1

print(f"\nMySQL Total Operations: {mysql_count}")
print(f"  - By operation: {dict(mysql_ops)}")
print(f"\nDynamoDB Total Operations: {dynamodb_count}")
print(f"  - By operation: {dict(dynamodb_ops)}")

if mysql_count == 150 and dynamodb_count == 150:
    print("\n✅ VALID: Both datasets have exactly 150 operations")
else:
    print("\n⚠️  WARNING: Datasets don't have 150 operations each!")
    print("   MySQL:", mysql_count, "(expected 150)")
    print("   DynamoDB:", dynamodb_count, "(expected 150)")
    print("   Continuing anyway...")

# Add database identifier to each result
print("\n📊 Merging datasets...")
for result in mysql_results:
    result['database'] = 'mysql'

for result in dynamodb_results:
    result['database'] = 'dynamodb'

# Combine into single flat array (as required by homework)
# This format makes it easy to analyze - single source for ALL analysis
combined = mysql_results + dynamodb_results

# Save combined results (flat array format)
with open('combined_results.json', 'w') as f:
    json.dump(combined, f, indent=2)

print(f"✅ Combined dataset saved to: combined_results.json")
print(f"   Total operations: {len(combined)} (150 MySQL + 150 DynamoDB)")

# Save verification results
verification = {
    'mysql_total': mysql_count,
    'dynamodb_total': dynamodb_count,
    'mysql_by_operation': dict(mysql_ops),
    'dynamodb_by_operation': dict(dynamodb_ops),
    'valid': mysql_count == 150 and dynamodb_count == 150
}

with open('data_verification.json', 'w') as f:
    json.dump(verification, f, indent=2)

print(f"✅ Verification results saved to: data_verification.json")

print("\n" + "=" * 60)
print("✅ Part 0 complete! Ready for performance analysis.")
print("=" * 60)

