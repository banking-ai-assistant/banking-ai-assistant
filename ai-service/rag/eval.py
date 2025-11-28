import os
import json
from typing import Dict, List

def load_all_templates_from_dir() -> List[Dict]:
    current_dir = os.path.dirname(os.path.abspath(__file__))
    templates_dir = os.path.join(current_dir, "templates")
    templates = []
    for filename in os.listdir(templates_dir):
        if filename.endswith(".json"):
            filepath = os.path.join(templates_dir, filename)
            with open(filepath, 'r', encoding='utf-8') as f:
                data = json.load(f)
                for template in data.get("templates", []):
                    template["type"] = data["type"] 
                    templates.append(template)
    return templates

def build_keywords_from_examples(templates: List[Dict]) -> Dict[str, List[str]]:
    block_phrases = {}
    
    for template in templates:
        structure_blocks = [b.strip() for b in template["structure"].split("→")]
        example = template["example"].lower()
        
        for block in structure_blocks:
            phrases = []
            
            if block == "Подтверждение получения":
                if "подтверждаем получение" in example:
                    phrases.append("подтверждаем получение")
                elif "подтверждаем" in example:
                    phrases.append("подтверждаем")
                    
            elif block == "Благодарность за обращение" or block == "Благодарность":
                if "благодарим за" in example:
                    phrases.append("благодарим за")
                elif "благодарим" in example:
                    phrases.append("благодарим")
                    
            elif block == "Признание проблемы":
                if "признаём" in example:
                    phrases.append("признаём")
                    
            elif block == "Процедура согласования" or block == "Процедура рассмотрения":
                if "рассмотрение" in example and ("процедура" in example or "согласование" in example or "регламент" in example):
                    phrases.append("рассмотрение")
                    
            elif block == "Правовая позиция":
                if "в соответствии с" in example or "согласно" in example:
                    phrases.append("в соответствии с")
                    
            elif block == "Сроки рассмотрения" or block == "Сроки предоставления" or block == "Сроки решения" or block == "Сроки":
                if "до" in example or "в течение" in example:
                    phrases.append("до")
                    phrases.append("в течение")
                    
            elif block == "Требуемые документы" or block == "Состав предоставляемых документов":
                if "документ" in example or "материал" in example:
                    phrases.append("документ")
                    phrases.append("материал")
                    
            elif block == "Контакты" or block == "Контакты для связи" or block == "Ответственное лицо":
                if "свяжемся" in example or "контакт" in example or "ответственный" in example:
                    phrases.append("свяжемся")
                    
            elif block == "Компенсация":
                if "компенсаци" in example or "предлагаем" in example and ("компенсация" in example or "возместим" in example):
                    phrases.append("компенсаци")
                    
            elif block == "Интерес к сотрудничеству":
                if "интерес" in example and ("сотрудничество" in example or "партнёрство" in example):
                    phrases.append("интерес")
                    
            elif block == "Дальнейшие шаги":
                if "изучим" in example or "свяжемся для обсуждения" in example:
                    phrases.append("изучим")
                    phrases.append("свяжемся для обсуждения")
                    
            elif block == "Ссылка на нормативный акт":
                if "указание банка россии" in example or "нормативный акт" in example:
                    phrases.append("указание банка россии")
                    phrases.append("нормативный акт")
            
            if phrases:
                if block not in block_phrases:
                    block_phrases[block] = []
                block_phrases[block].extend(phrases)
    
    for block in block_phrases:
        block_phrases[block] = list(set(block_phrases[block]))
    
    return block_phrases

def evaluate_response_by_examples(response: str, template: Dict, block_phrases: Dict[str, List[str]]) -> Dict:
    response_lower = response.lower()
    structure_blocks = [b.strip() for b in template["structure"].split("→")]
    
    matched = 0
    details = []
    
    for block in structure_blocks:
        phrases = block_phrases.get(block, [])
        found = False
        
        for phrase in phrases:
            if phrase in response_lower:
                found = True
                break
        
        if found:
            matched += 1
            details.append(f" {block}")
        else:
            details.append(f" {block}")
    
    score = matched / len(structure_blocks) if structure_blocks else 0
    return {
        "score": round(score, 2),
        "matched": matched,
        "total": len(structure_blocks),
        "details": details
    }

def evaluate_by_template(response: str, template: Dict) -> Dict:
    return evaluate_response_by_examples(response, template, BLOCK_PHRASES)

TEMPLATES = load_all_templates_from_dir()
BLOCK_PHRASES = build_keywords_from_examples(TEMPLATES)