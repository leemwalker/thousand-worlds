package interaction

// Dialogue Templates
var (
	GreetingTemplates = map[string][]string{
		"high_affection": {
			"Hello friend! It is good to see you!",
			"Greetings, {name}! How have you been?",
			"A pleasure to meet you here, {name}.",
		},
		"medium_affection": {
			"Greetings.",
			"Hello there.",
			"Good day.",
		},
		"low_affection": {
			"Oh, it's you.",
			"*nods curtly*",
			"What do you want?",
		},
	}

	TopicTemplates = map[string]string{
		"memory":     "I recently experienced {topic}.",
		"desire":     "I am feeling quite {topic}.",
		"small_talk": "Have you noticed the {topic} lately?",
	}

	ResponseTemplates = map[string]string{
		"agreeable":    "That sounds {adjective}! I agree completely.",
		"neutral":      "I see. That is interesting.",
		"disagreeable": "I don't think so. That sounds {adjective}.",
	}

	Adjectives = map[string][]string{
		"positive": {"wonderful", "exciting", "delightful", "fascinating"},
		"negative": {"unfortunate", "terrible", "boring", "dreadful"},
	}
)
