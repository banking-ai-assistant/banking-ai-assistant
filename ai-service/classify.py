from config import get_ai_client
import re
import json

class LetterClassifier:
    def __init__(self):
        self.letter_types = {
            'complaint': 'Жалоба',
            'information_request': 'Запрос информации',
            'regulatory': 'Регуляторный запрос',
            'partnership': 'Партнерское предложение',
            'approval_request': 'Запрос согласования',
            'notification': 'Уведомление'
        }
    
    def predict(self, text: str) -> dict:
        try:
            return self._ai_classify(text)
        except Exception as e:
            print(f"AI классификация не сработала: {e}")
            return self._smart_rule_based_classify(text)
    
    def _ai_classify(self, text: str) -> dict:
        client, model = get_ai_client()
        
        prompt = f"""
        Проанализируй это банковское письмо и верни полный анализ в JSON:

        "{text}"

        Верни JSON:
        {{
            "type": "complaint|information_request|regulatory|partnership|approval_request|notification",
            "urgency": "high|medium|low",
            "tone": "formal|neutral|emotional",
            "summary": "краткое содержание письма 1-2 предложения",
            "keywords": ["ключевое_слово1", "ключевое_слово2", "ключевое_слово3"]
        }}
        """
        
        response = client.responses.create(
            model=model,
            instructions="Ты - AI-аналитик банка. Анализируй письма и возвращай полный анализ в JSON формате.",
            input=prompt
        )
        
        result = self._parse_json_response(response.output_text)

        return {
            "type": result.get("type", "information_request"),
            "urgency": result.get("urgency", "medium"),
            "tone": result.get("tone", "neutral"),
            "summary": result.get("summary", "Письмо требует обработки"),
            "keywords": result.get("keywords", ["обработка"])
        }
    
    def _smart_rule_based_classify(self, text: str) -> dict:
        simple_result = self._simple_rules_classify(text)
        if self._is_complex_case(text, simple_result):
            return self._llm_enhance_classification(text, simple_result)
        
        return simple_result
    
    def _simple_rules_classify(self, text: str) -> dict:
        text_lower = text.lower()
        if any(word in text_lower for word in ['жалоб', 'претензи', 'недоволен', 'злюсь', 'возмущ']):
            return self._create_base_result('complaint', 'high', 'emotional')
        elif any(word in text_lower for word in ['фз', 'закон', 'регулятор', 'центробанк', 'надзор']):
            return self._create_base_result('regulatory', 'high', 'formal')
        elif any(word in text_lower for word in ['предоставьте', 'вышлите', 'документ', 'прошу', 'необходимо', 'запрос']):
            return self._create_base_result('information_request', 'medium', 'neutral')
        elif any(word in text_lower for word in ['сотрудничество', 'партнер', 'предлагаем', 'взаимодейств']):
            return self._create_base_result('partnership', 'low', 'neutral')
        elif any(word in text_lower for word in ['согласован', 'утвердите', 'одобрен', 'согласование']):
            return self._create_base_result('approval_request', 'medium', 'formal')
        else:
            return self._create_base_result('information_request', 'medium', 'neutral')
    
    def _is_complex_case(self, text: str, simple_result: dict) -> bool:

        is_long_text = len(text) > 150
        is_generic_type = simple_result['type'] in ['information_request', 'notification']
        has_generic_keywords = all(keyword in ['обработка', 'анализ', 'уведомление'] 
                                 for keyword in simple_result['keywords'])
        
        return is_long_text and (is_generic_type or has_generic_keywords)
    
    def _llm_enhance_classification(self, text: str, base_result: dict) -> dict:
        try:
            client, model = get_ai_client()
            
            prompt = f"""
            Проанализируй это письмо и выдели главное:
            
            "{text}"
            
            Основная тема: {base_result['type']}
            
            Верни JSON ТОЛЬКО с улучшенными полями:
            {{
                "summary": "короткое содержание 1 предложение",
                "keywords": ["3-5 самых важных слов из текста"]
            }}
            """
            
            response = client.responses.create(
                model=model,
                instructions="Выдели главную мысль и ключевые слова из письма. Верни только JSON.",
                input=prompt
            )
            
            result = self._parse_json_response(response.output_text)
            return {
                "type": base_result["type"],
                "urgency": base_result["urgency"],
                "tone": base_result["tone"],
                "summary": result.get("summary", base_result["summary"]),
                "keywords": result.get("keywords", base_result["keywords"])
            }
            
        except Exception as e:
            print(f"LLM парсинг не сработал: {e}")
            return base_result
    
    def _create_base_result(self, letter_type: str, urgency: str, tone: str) -> dict:
        """Создает базовый результат с стандартными значениями"""
        summary_map = {
            'complaint': 'Письмо содержит жалобу или претензию',
            'regulatory': 'Регуляторный запрос от надзорных органов', 
            'information_request': 'Запрос информации или документов',
            'partnership': 'Предложение о сотрудничестве',
            'approval_request': 'Запрос на согласование или утверждение',
            'notification': 'Информационное сообщение'
        }
        
        keywords_map = {
            'complaint': ['жалоба', 'претензия'],
            'regulatory': ['регулятор', 'закон'], 
            'information_request': ['запрос', 'документы'],
            'partnership': ['партнерство', 'сотрудничество'],
            'approval_request': ['согласование', 'утверждение'],
            'notification': ['уведомление', 'информирование']
        }
        
        return {
            "type": letter_type,
            "urgency": urgency,
            "tone": tone,
            "summary": summary_map.get(letter_type, 'Письмо требует обработки'),
            "keywords": keywords_map.get(letter_type, ['обработка'])
        }
    
    def _parse_json_response(self, text: str) -> dict:
        try:
            match = re.search(r'\{.*\}', text, re.DOTALL)
            if match:
                return json.loads(match.group(0))
            return {}
        except Exception as e:
            print(f"Ошибка парсинга JSON: {e}")
            return {}