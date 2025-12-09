export interface Skill {
    name: string;
    category: string;
    level: number;
    xp: number;
}

export interface SkillSheet {
    character_id: string;
    skills: Record<string, Skill>;
}
