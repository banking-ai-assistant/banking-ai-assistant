from fastapi import FastAPI
from ai import analyze, generate

app = FastAPI()

@app.post("/analyze")
async def analyze_endpoint(request: dict):
    return analyze(request["content"])

@app.post("/generate")
async def generate_endpoint(request: dict):
    return generate(request["content"], request["action"])

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8000)