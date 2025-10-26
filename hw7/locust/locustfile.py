from locust import HttpUser, task, between
import random
import json

class OrderUser(HttpUser):
    wait_time = between(0.1, 0.5)
    
    @task
    def create_sync_order(self):
        """
        Simulates a customer placing an order
        """
        order = {
            "customer_id": random.randint(1, 1000),
            "items": [
                {
                    "product_id": f"item-{random.randint(1, 100)}",
                    "quantity": random.randint(1, 5),
                    "price": round(random.uniform(10.0, 100.0), 2)
                }
            ]
        }
        
        # POST to /orders/sync endpoint
        with self.client.post(
            "/orders/sync",
            json=order,
            timeout=5.0, # Customer won't wait for more than 5 seconds
            catch_response=True
        ) as response:
            if response.status_code == 200:
                response.success()
            else:
                response.failure(f"Got status code {response.status_code}")

    @task
    def create_async_order(self):
        """Test async endpoint"""
        order = {
            "customer_id": random.randint(1, 1000),
            "items": [
                {
                    "product_id": f"item-{random.randint(1, 100)}",
                    "quantity": random.randint(1, 5),
                    "price": round(random.uniform(10.0, 100.0), 2)
                }
            ]
        }
        
        with self.client.post(
            "/orders/async",
            json=order,
            catch_response=True
        ) as response:
            if response.status_code == 202:
                response.success()
            else:
                response.failure(f"Got status code {response.status_code}")