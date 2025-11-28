class SemanticSearch:
    def find_best_templates(self, letter_type: str, keywords: List[str], tone: str) -> List[Dict]:
        templates = self.template_loader.get_templates_by_type(letter_type)
        
        scored_templates = []
        for template in templates:
            score = self._calculate_relevance_score(template, keywords, tone)
            scored_templates.append((score, template))
        
        scored_templates.sort(reverse=True)
        return [tpl for score, tpl in scored_templates[:2]]  
    
    def _calculate_relevance_score(self, template: Dict, keywords: List[str], tone: str) -> float:
        score = 0
        if template['tone'] == tone:
            score += 2

        template_text = template['name'] + " " + template['structure']
        for keyword in keywords:
            if keyword.lower() in template_text.lower():
                score += 1
        return score