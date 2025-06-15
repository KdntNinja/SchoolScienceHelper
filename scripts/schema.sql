-- Users
CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    username TEXT NOT NULL,
    created_at BIGINT NOT NULL,
    avatar_url TEXT
);

-- Projects
CREATE TABLE IF NOT EXISTS projects (
    id TEXT PRIMARY KEY,
    owner_id TEXT REFERENCES users(id),
    title TEXT NOT NULL,
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL,
    data JSONB
);

-- Quizzes
CREATE TABLE IF NOT EXISTS quizzes (
    id TEXT PRIMARY KEY,
    owner_id TEXT REFERENCES users(id),
    title TEXT NOT NULL,
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL
);

-- Questions (now supports all subjects)
CREATE TABLE IF NOT EXISTS questions (
    id TEXT PRIMARY KEY,
    quiz_id TEXT REFERENCES quizzes(id),
    subject TEXT NOT NULL,
    topic TEXT NOT NULL,
    prompt TEXT NOT NULL,
    choices TEXT[] NOT NULL,
    answer INT NOT NULL,
    explanation TEXT
);

-- Quiz Results
CREATE TABLE IF NOT EXISTS quiz_results (
    id TEXT PRIMARY KEY,
    quiz_id TEXT REFERENCES quizzes(id),
    user_id TEXT REFERENCES users(id),
    score INT NOT NULL,
    started_at BIGINT NOT NULL,
    ended_at BIGINT NOT NULL,
    answers INT[]
);

-- Revision/Resources
CREATE TABLE IF NOT EXISTS revision_resources (
    id TEXT PRIMARY KEY,
    owner_id TEXT REFERENCES users(id),
    type TEXT NOT NULL,
    topic TEXT,
    content TEXT NOT NULL,
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL
);

-- General Resources
CREATE TABLE IF NOT EXISTS resources (
    id TEXT PRIMARY KEY,
    owner_id TEXT REFERENCES users(id),
    type TEXT NOT NULL,
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL
);

-- Leaderboard
CREATE TABLE IF NOT EXISTS leaderboard (
    user_id TEXT REFERENCES users(id),
    username TEXT NOT NULL,
    score INT NOT NULL,
    streak INT NOT NULL,
    rank INT NOT NULL,
    PRIMARY KEY(user_id)
);

-- Anki Decks (for imported Anki .apkg files)
CREATE TABLE IF NOT EXISTS anki_decks (
    id TEXT PRIMARY KEY,
    owner_id TEXT REFERENCES users(id),
    name TEXT NOT NULL,
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL
);

-- Anki Cards (linked to imported decks)
CREATE TABLE IF NOT EXISTS anki_cards (
    id TEXT PRIMARY KEY,
    deck_id TEXT REFERENCES anki_decks(id),
    owner_id TEXT REFERENCES users(id),
    front TEXT NOT NULL,
    back TEXT NOT NULL,
    media JSONB,
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL
);
