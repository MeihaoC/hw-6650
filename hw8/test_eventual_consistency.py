"""
Test eventual consistency in DynamoDB
Creates cart, immediately reads it multiple times to catch consistency delays
"""

import requests
import time
from datetime import datetime
import statistics

ALB_URL = "http://shopping-cart-service-alb-1967658374.us-west-2.elb.amazonaws.com"

def test_write_read_consistency(iterations=20):
    """
    Test 1: Create cart, immediately read it
    Measures if cart is immediately available after creation
    """
    print("=" * 70)
    print("TEST 1: Write-Then-Read Consistency")
    print("=" * 70)
    print(f"Creating {iterations} carts and immediately reading them...\n")
    
    delays_observed = 0
    read_times = []
    
    for i in range(iterations):
        # Step 1: Create cart
        create_start = time.time()
        create_response = requests.post(
            f"{ALB_URL}/shopping-carts",
            json={"customer_id": 9000 + i}
        )
        create_time = (time.time() - create_start) * 1000
        
        if create_response.status_code != 201:
            print(f"  ❌ Test {i+1}: Failed to create cart")
            continue
        
        cart_id = create_response.json()["shopping_cart_id"]
        
        # Step 2: IMMEDIATELY read the cart (no delay)
        read_start = time.time()
        read_response = requests.get(f"{ALB_URL}/shopping-carts/{cart_id}")
        read_time = (time.time() - read_start) * 1000
        read_times.append(read_time)
        
        # Step 3: Check if cart was found
        if read_response.status_code == 404:
            print(f"  ⚠️  Test {i+1}: EVENTUAL CONSISTENCY DELAY OBSERVED!")
            print(f"      Cart {cart_id} not found immediately after creation")
            delays_observed += 1
            
            # Try again after a delay
            time.sleep(0.1)  # 100ms
            retry_response = requests.get(f"{ALB_URL}/shopping-carts/{cart_id}")
            if retry_response.status_code == 200:
                print(f"      ✅ Cart found after 100ms retry")
            else:
                print(f"      ❌ Cart still not found after 100ms")
        
        elif read_response.status_code == 200:
            print(f"  ✅ Test {i+1}: Cart immediately available (create: {create_time:.1f}ms, read: {read_time:.1f}ms)")
        else:
            print(f"  ❌ Test {i+1}: Unexpected status {read_response.status_code}")
        
        time.sleep(0.05)  # Small delay between tests
    
    print("\n" + "-" * 70)
    print(f"📊 SUMMARY:")
    print(f"   Total Tests: {iterations}")
    print(f"   Consistency Delays: {delays_observed} ({delays_observed/iterations*100:.1f}%)")
    print(f"   Immediate Success: {iterations - delays_observed} ({(iterations-delays_observed)/iterations*100:.1f}%)")
    if read_times:
        print(f"   Avg Read Time: {statistics.mean(read_times):.2f}ms")
    print("-" * 70 + "\n")
    
    return delays_observed

def test_concurrent_reads(cart_id, num_reads=10):
    """
    Test 2: Multiple rapid reads of the same cart
    Tests if repeated reads return consistent data
    """
    print("=" * 70)
    print("TEST 2: Concurrent Read Consistency")
    print("=" * 70)
    print(f"Reading cart {cart_id} {num_reads} times rapidly...\n")
    
    results = []
    
    for i in range(num_reads):
        start = time.time()
        response = requests.get(f"{ALB_URL}/shopping-carts/{cart_id}")
        elapsed = (time.time() - start) * 1000
        
        if response.status_code == 200:
            cart_data = response.json()
            item_count = len(cart_data.get("items", []))
            results.append({
                "iteration": i + 1,
                "status": "success",
                "item_count": item_count,
                "time_ms": elapsed
            })
            print(f"  Read {i+1}: ✅ Success, {item_count} items, {elapsed:.1f}ms")
        else:
            results.append({
                "iteration": i + 1,
                "status": "failed",
                "status_code": response.status_code,
                "time_ms": elapsed
            })
            print(f"  Read {i+1}: ❌ Failed with status {response.status_code}")
        
        time.sleep(0.01)  # 10ms between reads
    
    # Check for inconsistencies
    item_counts = [r["item_count"] for r in results if r["status"] == "success"]
    if len(set(item_counts)) > 1:
        print(f"\n  ⚠️  INCONSISTENCY DETECTED: Different item counts across reads!")
        print(f"      Item counts observed: {set(item_counts)}")
    else:
        print(f"\n  ✅ All reads returned consistent data")
    
    print("-" * 70 + "\n")

def test_write_read_write_consistency():
    """
    Test 3: Create cart → Add item → Read cart
    Tests if item additions are immediately visible
    """
    print("=" * 70)
    print("TEST 3: Write-Read-Write Consistency")
    print("=" * 70)
    print("Testing if item additions are immediately visible...\n")
    
    delays_observed = 0
    iterations = 10
    
    for i in range(iterations):
        # Step 1: Create cart
        create_resp = requests.post(
            f"{ALB_URL}/shopping-carts",
            json={"customer_id": 8000 + i}
        )
        
        if create_resp.status_code != 201:
            print(f"  Test {i+1}: Failed to create cart")
            continue
        
        cart_id = create_resp.json()["shopping_cart_id"]
        
        # Step 2: Add item to cart
        add_resp = requests.post(
            f"{ALB_URL}/shopping-carts/{cart_id}/items",
            json={"product_id": 5000 + i, "quantity": 3}
        )
        
        if add_resp.status_code != 204:
            print(f"  Test {i+1}: Failed to add item")
            continue
        
        # Step 3: IMMEDIATELY read cart to see if item is there
        read_resp = requests.get(f"{ALB_URL}/shopping-carts/{cart_id}")
        
        if read_resp.status_code == 200:
            cart_data = read_resp.json()
            items = cart_data.get("items", [])
            
            if len(items) == 0:
                print(f"  ⚠️  Test {i+1}: Item NOT immediately visible after add!")
                delays_observed += 1
                
                # Retry after delay
                time.sleep(0.1)
                retry_resp = requests.get(f"{ALB_URL}/shopping-carts/{cart_id}")
                if retry_resp.status_code == 200:
                    retry_items = retry_resp.json().get("items", [])
                    if len(retry_items) > 0:
                        print(f"       ✅ Item visible after 100ms retry")
                    else:
                        print(f"       ❌ Item still not visible")
            else:
                print(f"  ✅ Test {i+1}: Item immediately visible ({len(items)} items in cart)")
        else:
            print(f"  Test {i+1}: Failed to read cart")
        
        time.sleep(0.05)
    
    print("\n" + "-" * 70)
    print(f"📊 SUMMARY:")
    print(f"   Consistency Delays: {delays_observed}/{iterations}")
    print("-" * 70 + "\n")

def test_update_propagation_time():
    """
    Test 4: Measure how long it takes for updates to propagate
    """
    print("=" * 70)
    print("TEST 4: Update Propagation Time")
    print("=" * 70)
    print("Measuring propagation time for cart updates...\n")
    
    # Create a cart
    create_resp = requests.post(
        f"{ALB_URL}/shopping-carts",
        json={"customer_id": 7777}
    )
    cart_id = create_resp.json()["shopping_cart_id"]
    
    # Add an item
    requests.post(
        f"{ALB_URL}/shopping-carts/{cart_id}/items",
        json={"product_id": 9999, "quantity": 1}
    )
    
    # Now update the quantity by adding more
    print("  Adding 5 more of the same product...")
    update_time = time.time()
    
    requests.post(
        f"{ALB_URL}/shopping-carts/{cart_id}/items",
        json={"product_id": 9999, "quantity": 5}
    )
    
    # Poll until we see the updated quantity (should be 6 total)
    max_polls = 20
    poll_interval = 0.01  # 10ms
    
    for poll in range(max_polls):
        time.sleep(poll_interval)
        
        read_resp = requests.get(f"{ALB_URL}/shopping-carts/{cart_id}")
        if read_resp.status_code == 200:
            items = read_resp.json().get("items", [])
            for item in items:
                if item["product_id"] == 9999:
                    if item["quantity"] == 6:
                        propagation_time = (time.time() - update_time) * 1000
                        print(f"  ✅ Update visible after {propagation_time:.1f}ms")
                        print(f"     (Polling iteration: {poll + 1})")
                        return
                    else:
                        print(f"  Poll {poll + 1}: Old quantity still visible ({item['quantity']})")
    
    print("  ⚠️  Update not visible after {max_polls * poll_interval * 1000}ms")
    print("-" * 70 + "\n")

def main():
    print("\n" + "=" * 70)
    print("🔬 DynamoDB EVENTUAL CONSISTENCY TESTING")
    print("=" * 70)
    print(f"Target: {ALB_URL}")
    print(f"Time: {datetime.now()}")
    print("=" * 70 + "\n")
    
    # Run all tests
    print("Running 4 consistency tests...\n")
    
    # Test 1: Basic write-read
    delays = test_write_read_consistency(iterations=20)
    
    # Test 2: Concurrent reads (create a cart first)
    print("Setting up cart for concurrent read test...")
    setup_resp = requests.post(f"{ALB_URL}/shopping-carts", json={"customer_id": 9999})
    if setup_resp.status_code == 201:
        test_cart_id = setup_resp.json()["shopping_cart_id"]
        # Add some items
        requests.post(f"{ALB_URL}/shopping-carts/{test_cart_id}/items",
                     json={"product_id": 1111, "quantity": 2})
        requests.post(f"{ALB_URL}/shopping-carts/{test_cart_id}/items",
                     json={"product_id": 2222, "quantity": 3})
        test_concurrent_reads(test_cart_id, num_reads=10)
    
    # Test 3: Write-read-write
    test_write_read_write_consistency()
    
    # Test 4: Update propagation
    test_update_propagation_time()
    
    print("\n" + "=" * 70)
    print("🎯 TESTING COMPLETE")
    print("=" * 70)
    print("\nKey Findings:")
    if delays == 0:
        print("✅ No eventual consistency delays observed in write-read test")
        print("   This is expected for:")
        print("   - Single-region DynamoDB")
        print("   - Light load")
        print("   - Modern AWS infrastructure")
    else:
        print(f"⚠️  {delays} consistency delays observed!")
        print("   This demonstrates eventual consistency in action")
    
    print("\nNote: DynamoDB's eventual consistency is typically <100ms")
    print("      Under light load, you may not observe delays at all.")
    print("\n" + "=" * 70 + "\n")

if __name__ == "__main__":
    main()