import uuid
from typing import List, Dict
from config import get_ai_client
from rag.template_loader import TemplateLoader
from classifier import LetterClassifier

class RAGEngine:
    def __init__(self):
        self.template_loader = TemplateLoader()
        self.classifier = LetterClassifier()  
        self.client, self.model = get_ai_client()
    
    def generate_response_options(self, content: str, action: str) -> List[Dict]:
        analysis = self.classifier.predict(content)
        
        templates = self.template_loader.get_templates_by_type(analysis['type'])
        if not templates:
            templates = self._get_fallback_templates(analysis['type'])
        
        options = []
        for template in templates[:2]:
            try:
                response_content = self._generate_with_template(
                    template, analysis, content, action
                )
                
                option = {
                    "id": str(uuid.uuid4())[:8],
                    "content": response_content,
                    "style": template['style'],
                    "summary": template['name']
                }
                options.append(option)
                
            except Exception as e:
                print(f"Ошибка генерации: {e}")
                continue
        
        return options if options else self._get_fallback_options()
    
    def _generate_with_template(self, template: Dict, analysis: Dict, content: str, action: str) -> str:
        prompt = f"""
        Сгенерируй ответ банка на письмо:

        ПИСЬМО: "{content}"

        Используй шаблон:
        - Структура: {template['structure']}
        - Стиль: {template['style']}
        - Пример: {template['example']}

        Тип письма: {analysis['type']}
        Действие: {action}
        """
        
        response = self.client.responses.create(
            model=self.model,
            instructions=f"Сгенерируй профессиональный ответ в стиле {template['style']}",
            input=prompt
        )
        
        return response.output_text
    
    def _get_fallback_templates(self, letter_type: str) -> List[Dict]:
        return [{
            "id": "general",
            "name": "Общий ответ",
            "structure": "Приветствие → Ответ → Контакты",
            "example": "Уважаемый клиент! В ответ на ваше обращение...",
            "style": "neutral",
            "tone": "neutral"
        }]
    
    def _get_fallback_options(self) -> List[Dict]:
        return [{
            "id": "fallback",
            "content": "Уважаемый клиент! Благодарим за обращение. Ваше письмо находится в обработке.",
            "style": "neutral",
            "summary": "Ответ об обработке"
        }]