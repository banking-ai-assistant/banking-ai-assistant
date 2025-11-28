import json
import os
from typing import Dict, List

class TemplateLoader:
    def __init__(self, templates_dir: str = "rag/templates"):
        self.templates_dir = templates_dir
        self.templates_cache = {}
    
    def load_all_templates(self) -> Dict:
        """Загружает все шаблоны из папки"""
        if self.templates_cache:
            return self.templates_cache
        
        templates = {}
        
        for filename in os.listdir(self.templates_dir):
            if filename.endswith('.json'):
                filepath = os.path.join(self.templates_dir, filename)
                with open(filepath, 'r', encoding='utf-8') as f:
                    template_data = json.load(f)
                    templates[template_data['type']] = template_data
        
        self.templates_cache = templates
        return templates
    
    def get_templates_by_type(self, letter_type: str) -> List:
        templates = self.load_all_templates()
        return templates.get(letter_type, {}).get('templates', [])
    
    def get_template_by_id(self, template_id: str) -> Dict:
        templates = self.load_all_templates()
        
        for letter_type, data in templates.items():
            for template in data.get('templates', []):
                if template['id'] == template_id:
                    return template
        
        return None