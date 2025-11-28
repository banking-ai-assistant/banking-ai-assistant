import uuid
from typing import List, Dict
from config import get_ai_client
from rag.template_loader import TemplateLoader
from classify import LetterClassifier

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
      improved_draft = self._generate_improved_draft(content, action, analysis)
      if improved_draft:
          options.append({
              "id": str(uuid.uuid4())[:8],
              "content": improved_draft,
              "style": "professional",
              "summary": "Профессиональная версия вашего черновика"
          })
      
      if templates:
          template_response = self._generate_with_template(
              templates[0], analysis, content, action
          )
          if template_response:
              options.append({
                  "id": str(uuid.uuid4())[:8],
                  "content": template_response,
                  "style": templates[0]['style'],
                  "summary": templates[0]['name']
              })

      while len(options) < 2:
          options.append(self._get_fallback_option(content, action, len(options)))
      
      return options
    
    def _generate_improved_draft(self, original_letter: str, user_draft: str, analysis: Dict, template: Dict = None) -> str:
        template_info = ""
        if template:
            template_info = f"""
            РЕКОМЕНДУЕМЫЙ ШАБЛОН:
            - Структура: {template['structure']}
            - Стиль: {template['style']}
            """
        
        prompt = f"""
        ИСХОДНОЕ ПИСЬМО КЛИЕНТА:
        "{original_letter}"

        ЧЕРНОВИК ОТВЕТА ОТ СОТРУДНИКА:
        "{user_draft}"

        ТИП ПИСЬМА: {analysis['type']}
        {template_info}

        Задача: Переработай черновик в профессиональный ответ банка.
        """
        
        try:
            instructions = "Улучши черновик сотрудника, сделав его профессиональным банковским ответом."
            if template:
                instructions += f" Учти структуру: {template['structure']}."
                
            response = self.client.responses.create(
                model=self.model,
                instructions=instructions,
                input=prompt
            )
            return response.output_text
        except Exception as e:
            print(f"Ошибка черновика: {e}")
            return ""
    
    def _generate_with_template(self, template: Dict, analysis: Dict, content: str, action: str) -> str:
        prompt = f"""
        ИСХОДНОЕ ПИСЬМО:
        "{content}"

        ЧЕРНОВИК СОТРУДНИКА (намерение):
        "{action}"

        ШАБЛОН:
        - Название: {template['name']}
        - Структура: {template['structure']}
        - Стиль: {template['style']}
        - Пример: {template['example']}

        Задача: Создай ответ по шаблону, который реализует намерение из черновика.
        """
        
        try:
            response = self.client.responses.create(
                model=self.model,
                instructions=f"Сгенерируй ответ в стиле {template['style']}, реализующий намерение из черновика сотрудника.",
                input=prompt
            )
            return response.output_text
        except Exception as e:
            print(f"Ошибка генерации по шаблону: {e}")
            return ""
    
    def _get_fallback_templates(self, letter_type: str) -> List[Dict]:
        return [{
            "id": "general",
            "name": "Общий ответ",
            "structure": "Приветствие → Ответ → Контакты",
            "example": "Уважаемый клиент! Рассмотрим позже...",
            "style": "neutral",
            "tone": "neutral"
        }]
    
    def _get_fallback_option(self, content: str, action: str, option_index: int) -> Dict:
        fallbacks = [
            {
                "id": str(uuid.uuid4())[:8],
                "content": f"Уважаемый клиент! {action}",
                "style": "fallback",
                "summary": "Ответ на основе вашего черновика"
            },
            {
                "id": str(uuid.uuid4())[:8],
                "content": "Уважаемый клиент! Благодарим за обращение. Мы рассмотрим ваш вопрос в ближайшее время.",
                "style": "neutral", 
                "summary": "Стандартный ответ"
            }
        ]
        return fallbacks[option_index % len(fallbacks)]