DROP TABLE IF EXISTS interview_answers;
DROP TABLE IF EXISTS world_configurations;
DROP TABLE IF EXISTS world_interviews;
DROP TYPE IF EXISTS interview_status;

CREATE TYPE interview_status AS ENUM ('not_started', 'in_progress', 'completed');

CREATE TABLE world_interviews (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    status interview_status NOT NULL DEFAULT 'not_started',
    current_question_index INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(user_id)
);

CREATE INDEX idx_interviews_user ON world_interviews(user_id);

CREATE TABLE interview_answers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    interview_id UUID NOT NULL REFERENCES world_interviews(id) ON DELETE CASCADE,
    question_index INT NOT NULL,
    answer_text TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(interview_id, question_index)
);

CREATE INDEX idx_answers_interview ON interview_answers(interview_id);

CREATE TABLE world_configurations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    interview_id UUID NOT NULL REFERENCES world_interviews(id) ON DELETE CASCADE,
    world_id UUID REFERENCES worlds(id) ON DELETE SET NULL,
    created_by UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    configuration JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_world_configs_interview ON world_configurations(interview_id);
CREATE INDEX idx_world_configs_created_by ON world_configurations(created_by);
