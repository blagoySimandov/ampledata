from google_search_results_model import GoogleSearchResults
import requests
import json
import os
from abc import ABC, abstractmethod


class WebSearcher(ABC):
    @abstractmethod
    def search(self, query: str) -> GoogleSearchResults:
        pass


class SerperWebSearcher(WebSearcher):
    def search(self, query: str) -> GoogleSearchResults:
        url = "https://google.serper.dev/search"
        payload = json.dumps({"q": query})
        headers = {
            "X-API-KEY": os.getenv("SERPER_API_KEY"),
            "Content-Type": "application/json",
        }
        response = requests.request("POST", url, headers=headers, data=payload)
        return response.json()
