-- GCSE Science DB Schema

CREATE TABLE IF NOT EXISTS specs (
    id SERIAL PRIMARY KEY,
    board TEXT NOT NULL,
    tier TEXT NOT NULL,
    subject TEXT NOT NULL,
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    UNIQUE(board, tier, subject, title)
);

CREATE TABLE IF NOT EXISTS papers (
    id SERIAL PRIMARY KEY,
    board TEXT NOT NULL,
    tier TEXT NOT NULL,
    year INT NOT NULL,
    subject TEXT NOT NULL,
    url TEXT NOT NULL,
    UNIQUE(board, tier, year, subject, url)
);

CREATE TABLE IF NOT EXISTS questions (
    id SERIAL PRIMARY KEY,
    board TEXT NOT NULL,
    tier TEXT NOT NULL,
    subject TEXT NOT NULL,
    topic TEXT NOT NULL,
    question TEXT NOT NULL,
    answer TEXT NOT NULL,
    UNIQUE(board, tier, subject, topic, question)
);

CREATE TABLE IF NOT EXISTS revision (
    id SERIAL PRIMARY KEY,
    board TEXT NOT NULL,
    tier TEXT NOT NULL,
    subject TEXT NOT NULL,
    topic TEXT NOT NULL,
    content TEXT NOT NULL,
    UNIQUE(board, tier, subject, topic)
);
