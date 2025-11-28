from classifier import LetterClassifier
from rag.engine import RAGEngine

def analyze(content: str) -> dict:
    classifier = LetterClassifier()
    return classifier.predict(content)

def generate(content: str, action: str) -> dict:
    rag_engine = RAGEngine()
    options = rag_engine.generate_response_options(content, action)
    return {
        "options": options
    }