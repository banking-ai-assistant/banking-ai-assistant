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
        try:
            client, model = get_ai_client()
            
            prompt = f"""
            Текст письма: "{text}"
            """
            
            response = client.responses.create(
                model=model,
                instructions="""
                Ты - AI-аналитик банка. Проанализируй письмо и верни ТОЛЬКО JSON без каких-либо других слов.
                
                Строго соблюдай этот формат:
                {
                    "type": "complaint|information_request|regulatory|partnership|approval_request|notification",
                    "urgency": "high|medium|low", 
                    "tone": "formal|neutral|emotional",
                    "summary": "краткое содержание 1-2 предложения",
                    "keywords": ["слово1", "слово2", "слово3"]
                }
                
                НИКАКИХ пояснений, НИКАКИХ markdown, ТОЛЬКО чистый JSON.
                """,
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
            
        except Exception as e:
            print(f"Ошибка в классификации: {e}")
            raise
    
    def _parse_json_response(self, text: str) -> dict:
        try:
            cleaned = text.strip()

            if cleaned.startswith('```json'):
                cleaned = cleaned[7:].strip()
            elif cleaned.startswith('```'):
                cleaned = cleaned[3:].strip()
            
            if cleaned.endswith('```'):
                cleaned = cleaned[:-3].strip()

            result = json.loads(cleaned)
            return result
            
        except json.JSONDecodeError as e:
            print(f"JSONDecodeError: {e}")
            
            # Пробуем найти JSON объект в тексте
            return self._extract_json_with_regex(text)
        except Exception as e:
            print(f"Неизвестная ошибка: {e}")
            return {}
    
    def _extract_json_with_regex(self, text: str) -> dict:
        try:
            json_pattern = r'\{[^{}]*\{[^{}]*\}[^{}]*\}|\{[^{}]*\}'
            matches = re.findall(json_pattern, text, re.DOTALL)
            
            if matches:
                json_str = max(matches, key=len)
                return json.loads(json_str)
            return self._extract_fields_with_regex(text)
            
        except Exception as e:
            print(f"Regex extraction failed: {e}")
            return {}
    
    def _extract_fields_with_regex(self, text: str) -> dict:
        result = {}

        type_pattern = r'"type"\s*:\s*"([^"]+)"'
        type_match = re.search(type_pattern, text)
        if type_match:
            result["type"] = type_match.group(1)
        else:
            type_pattern2 = r'type\s*:\s*(\w+)'
            type_match2 = re.search(type_pattern2, text)
            if type_match2:
                result["type"] = type_match2.group(1)
        
        urgency_pattern = r'"urgency"\s*:\s*"([^"]+)"'
        urgency_match = re.search(urgency_pattern, text)
        if urgency_match:
            result["urgency"] = urgency_match.group(1)
        else:
            urgency_pattern2 = r'urgency\s*:\s*(\w+)'
            urgency_match2 = re.search(urgency_pattern2, text)
            if urgency_match2:
                result["urgency"] = urgency_match2.group(1)
        
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
        
        generic_keywords = ['обработка', 'анализ', 'уведомление', 'документы', 'запрос']
        has_generic_keywords = all(keyword in generic_keywords for keyword in simple_result['keywords'])
        
        return is_long_text and (is_generic_type or has_generic_keywords)
    
    def _llm_enhance_classification(self, text: str, base_result: dict) -> dict:
        try:
            client, model = get_ai_client()
            
            prompt = f"""
            Текст письма: "{text}"
            """
            
            response = client.responses.create(
                model=model,
                instructions=f"""
                Улучши анализ этого письма. Основная тема: {base_result['type']}
                Верни ТОЛЬКО JSON:
                {{
                    "summary": "короткое содержание 1 предложение",
                    "keywords": ["3-5 самых важных слов из текста"]
                }}
                """,
                input=prompt
            )
            
            print(f"🦄 LLM улучшение (сырое): {response.output_text}")
            result = self._parse_json_response(response.output_text)
            
            return {
                "type": base_result["type"],
                "urgency": base_result["urgency"],
                "tone": base_result["tone"],
                "summary": result.get("summary", base_result["summary"]),
                "keywords": result.get("keywords", base_result["keywords"])
            }
            
        except Exception as e:
            print(f"LLM улучшение не сработало: {e}")
            return base_result
    
    def _create_base_result(self, letter_type: str, urgency: str, tone: str) -> dict:
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
