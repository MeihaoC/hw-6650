"""
Performance Test Script for Shopping Cart Service
Tests 150 operations: 50 create, 50 add items, 50 get cart
Saves results to dynamodb_test_results.json or mysql_test_results.json
Automatically detects database type from health check
"""

import json
import time
import requests
from datetime import datetime
from typing import List, Dict
import sys

API_URL = "http://shopping-cart-service-alb-1967658374.us-west-2.elb.amazonaws.com"

def test_create_cart(customer_id: int) -> Dict:
    """Test POST /shopping-carts"""
    start_time = time.time()
    try:
        response = requests.post(
            f"{API_URL}/shopping-carts",
            json={"customer_id": customer_id},
            headers={"Content-Type": "application/json"},
            timeout=10
        )
        elapsed = (time.time() - start_time) * 1000  # Convert to ms
        
        return {
            "operation": "create_cart",
            "response_time": round(elapsed, 2),
            "success": response.status_code == 201,
            "status_code": response.status_code,
            "timestamp": datetime.utcnow().isoformat() + "Z",
            "cart_id": response.json().get("shopping_cart_id") if response.status_code == 201 else None
        }
    except Exception as e:
        elapsed = (time.time() - start_time) * 1000
        return {
            "operation": "create_cart",
            "response_time": round(elapsed, 2),
            "success": False,
            "status_code": 0,
            "timestamp": datetime.utcnow().isoformat() + "Z",
            "error": str(e)
        }

def test_add_item(cart_id, product_id: int, quantity: int) -> Dict:
    """Test POST /shopping-carts/{id}/items"""
    start_time = time.time()
    try:
        response = requests.post(
            f"{API_URL}/shopping-carts/{cart_id}/items",
            json={"product_id": product_id, "quantity": quantity},
            headers={"Content-Type": "application/json"},
            timeout=10
        )
        elapsed = (time.time() - start_time) * 1000
        
        return {
            "operation": "add_items",
            "response_time": round(elapsed, 2),
            "success": response.status_code == 204,
            "status_code": response.status_code,
            "timestamp": datetime.utcnow().isoformat() + "Z"
        }
    except Exception as e:
        elapsed = (time.time() - start_time) * 1000
        return {
            "operation": "add_items",
            "response_time": round(elapsed, 2),
            "success": False,
            "status_code": 0,
            "timestamp": datetime.utcnow().isoformat() + "Z",
            "error": str(e)
        }

def test_get_cart(cart_id) -> Dict:
    """Test GET /shopping-carts/{id}"""
    start_time = time.time()
    try:
        response = requests.get(
            f"{API_URL}/shopping-carts/{cart_id}",
            timeout=10
        )
        elapsed = (time.time() - start_time) * 1000
        
        return {
            "operation": "get_cart",
            "response_time": round(elapsed, 2),
            "success": response.status_code == 200,
            "status_code": response.status_code,
            "timestamp": datetime.utcnow().isoformat() + "Z"
        }
    except Exception as e:
        elapsed = (time.time() - start_time) * 1000
        return {
            "operation": "get_cart",
            "response_time": round(elapsed, 2),
            "success": False,
            "status_code": 0,
            "timestamp": datetime.utcnow().isoformat() + "Z",
            "error": str(e)
        }

def main():
    print("🚀 Starting Performance Test")
    print("=" * 50)
    print(f"API URL: {API_URL}")
    print("Test: 50 create, 50 add items, 50 get cart")
    print("=" * 50)
    print()
    
    results: List[Dict] = []
    created_carts: List[str] = []  # DynamoDB uses string UUIDs
    
    # Step 1: Create 50 carts
    print("📝 Creating 50 shopping carts...")
    for i in range(50):
        result = test_create_cart(customer_id=1000 + i)
        results.append(result)
        if result["success"] and result.get("cart_id"):
            created_carts.append(result["cart_id"])
        
        if (i + 1) % 10 == 0:
            print(f"   Created {i + 1}/50 carts...")
    
    print(f"✅ Created {len(created_carts)} carts")
    print()
    
    # Step 2: Add items to 50 carts (using first 50 carts)
    print("🛒 Adding items to 50 carts...")
    for i in range(50):
        if i < len(created_carts):
            cart_id = created_carts[i]
            product_id = 100 + (i % 10)  # Vary product IDs
            quantity = (i % 5) + 1  # Quantity 1-5
            result = test_add_item(cart_id, product_id, quantity)
            results.append(result)
        
        if (i + 1) % 10 == 0:
            print(f"   Added items to {i + 1}/50 carts...")
    
    print("✅ Added items to carts")
    print()
    
    # Step 3: Get 50 carts
    print("📖 Retrieving 50 carts...")
    for i in range(50):
        if i < len(created_carts):
            cart_id = created_carts[i]
            result = test_get_cart(cart_id)
            results.append(result)
        
        if (i + 1) % 10 == 0:
            print(f"   Retrieved {i + 1}/50 carts...")
    
    print("✅ Retrieved carts")
    print()
    
    # Save results - determine output file based on API response
    # Check if using DynamoDB or MySQL by health check
    try:
        health = requests.get(f"{API_URL}/health", timeout=5).json()
        db_type = health.get("database", "MySQL").lower()
        output_file = "dynamodb_test_results.json" if "dynamodb" in db_type else "mysql_test_results.json"
    except:
        # Default to DynamoDB if can't determine
        output_file = "dynamodb_test_results.json"
    with open(output_file, 'w') as f:
        json.dump(results, f, indent=2)
    
    # Print summary
    print("=" * 50)
    print("📊 Test Summary")
    print("=" * 50)
    
    total_ops = len(results)
    successful_ops = sum(1 for r in results if r["success"])
    
    create_ops = [r for r in results if r["operation"] == "create_cart"]
    add_ops = [r for r in results if r["operation"] == "add_items"]
    get_ops = [r for r in results if r["operation"] == "get_cart"]
    
    print(f"Total Operations: {total_ops}")
    print(f"Successful: {successful_ops} ({successful_ops/total_ops*100:.1f}%)")
    print()
    
    if create_ops:
        create_times = [r["response_time"] for r in create_ops if r["success"]]
        if create_times:
            print(f"Create Cart:")
            print(f"  Avg: {sum(create_times)/len(create_times):.2f} ms")
            print(f"  Min: {min(create_times):.2f} ms")
            print(f"  Max: {max(create_times):.2f} ms")
    
    if add_ops:
        add_times = [r["response_time"] for r in add_ops if r["success"]]
        if add_times:
            print(f"Add Items:")
            print(f"  Avg: {sum(add_times)/len(add_times):.2f} ms")
            print(f"  Min: {min(add_times):.2f} ms")
            print(f"  Max: {max(add_times):.2f} ms")
    
    if get_ops:
        get_times = [r["response_time"] for r in get_ops if r["success"]]
        if get_times:
            print(f"Get Cart:")
            print(f"  Avg: {sum(get_times)/len(get_times):.2f} ms")
            print(f"  Min: {min(get_times):.2f} ms")
            print(f"  Max: {max(get_times):.2f} ms")
    
    print()
    print(f"✅ Results saved to: {output_file}")
    print(f"📝 Total test duration: Must complete within 5 minutes (per homework)")

if __name__ == "__main__":
    try:
        main()
    except KeyboardInterrupt:
        print("\n⚠️  Test interrupted by user")
        sys.exit(1)
    except Exception as e:
        print(f"\n❌ Test failed with error: {e}")
        sys.exit(1)

