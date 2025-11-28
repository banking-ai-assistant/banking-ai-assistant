from classifier import LetterClassifier

def analyze(content: str) -> dict:
    classifier = LetterClassifier()
    return classifier.predict(content)

def generate(content: str, action: str) -> dict:
    return {
        "options": [
            {
                "id": "1",
                "content": "Заглушка",
                "style": "Заглушка", 
                "summary": "Заглушка"
            }
        ]
    }