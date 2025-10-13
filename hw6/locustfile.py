from locust import task, between
from locust.contrib.fasthttp import FastHttpUser
import random

class ProductSearchUser(FastHttpUser):
    """
    Load test user for product search service.
    Uses FastHttpUser for better performance.
    """
    
    # Minimal wait time between requests for maximum load
    # wait_time = between(0.1, 0.5)  # 0.1-0.5 seconds between requests
    
    # Common search terms that should return results
    search_queries = [
        "electronics",
        "books",
        "home",
        "clothing",
        "sports",
        "toys",
        "garden",
        "automotive",
        "alpha",
        "beta",
        "gamma",
        "delta",
        "epsilon",
        "product"
    ]
    
    @task
    def search_products(self):
        """
        Search for products using various query terms.
        This simulates the main load on the service.
        """
        # Randomly select a search term
        query = random.choice(self.search_queries)
        
        # Perform GET request to /products/search endpoint
        with self.client.get(
            f"/products/search?q={query}",
            catch_response=True,
            name="/products/search"  # Groups all searches together in stats
        ) as response:
            if response.status_code == 200:
                try:
                    json_data = response.json() # Converts response body from JSON string to Python dictionary
                    # Verify response structure
                    if "products" in json_data and "total_found" in json_data:
                        response.success()
                    else:
                        response.failure("Invalid response structure")
                except Exception as e:
                    response.failure(f"Failed to parse JSON: {str(e)}")
            else:
                response.failure(f"Got status code {response.status_code}")