import os
import requests
from dotenv import load_dotenv
from openai import OpenAI

def setup_env():
    try:
        env_url = "https://storage.yandexcloud.net/ycpub/maikeys/.env"
        response = requests.get(env_url)
        if response.status_code == 200:
            with open('.env', 'w') as f:
                f.write(response.text)
    except:
        pass
    
    load_dotenv()

def get_ai_client():
    setup_env()
    folder_id = os.environ.get('folder_id')
    api_key = os.environ.get('api_key')
    
    model = f"gpt://{folder_id}/yandexgpt-lite"
    
    client = OpenAI(
        base_url="https://rest-assistant.api.cloud.yandex.net/v1",
        api_key=api_key,
        project=folder_id
    )
    
    return client, model