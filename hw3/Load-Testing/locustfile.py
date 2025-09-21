from locust import FastHttpUser, task, between
import random

class AlbumStoreUser(FastHttpUser):
    wait_time = between(1, 3)  # Wait 1-3 seconds between tasks
    
    @task(3)  # Weight of 3 for GET requests
    def get_albums(self):
        self.client.get("/albums")
    
    @task(3)  # Another GET endpoint
    def get_album_by_id(self):
        album_id = random.choice(["1", "2", "3"])
        self.client.get(f"/albums/{album_id}")
    
    @task(1)  # Weight of 1 for POST (creates roughly 3:1 GET:POST ratio)
    def post_album(self):
        new_album = {
            "id": str(random.randint(100, 999)),
            "title": f"Test Album {random.randint(1, 100)}",
            "artist": f"Test Artist {random.randint(1, 50)}",
            "price": round(random.uniform(10.99, 99.99), 2)
        }
        self.client.post("/albums", json=new_album)