from locust import FastHttpUser, task, between
import random

# Test 2: FastHttpUser (more efficient)
class FastProductAPIUser(FastHttpUser):
    """
    FastHttpUser uses connection pooling and is more efficient.
    Better for high load testing (100+ users).
    """
    wait_time = between(1, 3)
    
    def on_start(self):
        """Called when a user starts"""
        for i in range(1, 4):
            product_id = random.randint(1000, 9999)
            self.create_product(product_id)
    
    def create_product(self, product_id):
        """Helper method to create a product"""
        product_data = {
            "product_id": product_id,
            "sku": f"SKU-{product_id}",
            "manufacturer": random.choice([
                "Acme Corporation",
                "Tech Industries",
                "Global Manufacturing",
                "Prime Products"
            ]),
            "category_id": random.randint(1, 100),
            "weight": random.randint(100, 5000),
            "some_other_id": random.randint(1, 1000)
        }
        
        self.client.post(
            f"/products/{product_id}/details",
            json=product_data,
            name="/products/[id]/details (POST)"
        )
    
    @task(5)
    def get_product(self):
        product_id = random.randint(1000, 9999)
        with self.client.get(
            f"/products/{product_id}",
            catch_response=True,
            name="/products/[id] (GET)"
        ) as response:
            if response.status_code == 404:
                response.success()
    
    @task(1)
    def create_new_product(self):
        product_id = random.randint(10000, 99999)
        self.create_product(product_id)
    
    @task(3)
    def health_check(self):
        self.client.get("/health", name="/health (GET)")