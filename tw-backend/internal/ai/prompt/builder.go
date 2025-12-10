package prompt

import (
	"bytes"
	"fmt"
	"mud-platform-backend/internal/character"
	"mud-platform-backend/internal/npc/memory"
	"mud-platform-backend/internal/npc/personality"
	"mud-platform-backend/internal/npc/relationship"
	"text/template"
)

// PromptBuilder constructs the dialogue prompt
type PromptBuilder struct {
	npcName      string
	age          int
	species      string
	occupation   string
	personality  *personality.Personality
	mood         *personality.Mood
	desire       string
	urgency      float64
	attributes   character.Attributes
	location     string
	timeOfDay    string
	weather      string
	entities     string
	speakerName  string
	relationship *relationship.Relationship
	memories     []memory.Memory
	driftMetrics *relationship.DriftMetrics
	baseline     relationship.BehavioralProfile
	currentBehav relationship.BehavioralProfile
	topic        string
	input        string
}

// NewPromptBuilder creates a new builder
func NewPromptBuilder() *PromptBuilder {
	return &PromptBuilder{}
}

func (pb *PromptBuilder) WithNPC(name string, age int, species, occupation string, p *personality.Personality, mood *personality.Mood, desire string, urgency float64, attr character.Attributes) *PromptBuilder {
	pb.npcName = name
	pb.age = age
	pb.species = species
	pb.occupation = occupation
	pb.personality = p
	pb.mood = mood
	pb.desire = desire
	pb.urgency = urgency
	pb.attributes = attr
	return pb
}

func (pb *PromptBuilder) WithContext(location, timeOfDay, weather, entities string) *PromptBuilder {
	pb.location = location
	pb.timeOfDay = timeOfDay
	pb.weather = weather
	pb.entities = entities
	return pb
}

func (pb *PromptBuilder) WithSpeaker(name string, rel *relationship.Relationship) *PromptBuilder {
	pb.speakerName = name
	pb.relationship = rel
	return pb
}

func (pb *PromptBuilder) WithMemories(mems []memory.Memory) *PromptBuilder {
	pb.memories = mems
	return pb
}

func (pb *PromptBuilder) WithDrift(drift *relationship.DriftMetrics, baseline, current relationship.BehavioralProfile) *PromptBuilder {
	pb.driftMetrics = drift
	pb.baseline = baseline
	pb.currentBehav = current
	return pb
}

func (pb *PromptBuilder) WithConversation(topic, input string) *PromptBuilder {
	pb.topic = topic
	pb.input = input
	return pb
}

// Build generates the final prompt string
func (pb *PromptBuilder) Build() (string, error) {
	data := map[string]interface{}{
		"NPCName":      pb.npcName,
		"Age":          pb.age,
		"Species":      pb.species,
		"Occupation":   pb.occupation,
		"Personality":  buildPersonalitySection(pb.personality),
		"State":        buildStateSection(pb.mood, pb.desire, pb.urgency, pb.attributes),
		"Context":      fmt.Sprintf("- Location: %s\n- Time: %s, %s\n- Nearby: %s", pb.location, pb.timeOfDay, pb.weather, pb.entities),
		"SpeakerName":  pb.speakerName,
		"Relationship": buildRelationshipSection(pb.relationship),
		"Memories":     buildMemoriesSection(pb.memories),
		"Drift":        buildDriftSection(pb.driftMetrics, pb.baseline, pb.currentBehav),
		"Topic":        pb.topic,
		"Input":        pb.input,
	}

	tmpl, err := template.New("prompt").Parse(BasePromptTemplate)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}
