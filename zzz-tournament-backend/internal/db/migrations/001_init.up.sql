CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Users table
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    rating INTEGER DEFAULT 1000,
    wins INTEGER DEFAULT 0,
    losses INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Heroes table
CREATE TABLE heroes (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    element VARCHAR(50) NOT NULL,
    rarity VARCHAR(20) NOT NULL,
    role VARCHAR(50) NOT NULL,
    description TEXT,
    image_url VARCHAR(500),
    is_active BOOLEAN DEFAULT true
);

-- Rooms table
CREATE TABLE rooms (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    host_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    max_players INTEGER DEFAULT 8,
    current_count INTEGER DEFAULT 1,
    status VARCHAR(20) DEFAULT 'waiting',
    is_private BOOLEAN DEFAULT false,
    password VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Room participants table
CREATE TABLE room_participants (
    room_id INTEGER REFERENCES rooms(id) ON DELETE CASCADE,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (room_id, user_id)
);

-- Tournaments table
CREATE TABLE tournaments (
    id SERIAL PRIMARY KEY,
    room_id INTEGER REFERENCES rooms(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    status VARCHAR(20) DEFAULT 'created',
    bracket JSONB,
    winner_id INTEGER REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Matches table
CREATE TABLE matches (
    id SERIAL PRIMARY KEY,
    tournament_id INTEGER REFERENCES tournaments(id) ON DELETE CASCADE,
    round INTEGER NOT NULL,
    player1_id INTEGER REFERENCES users(id),
    player2_id INTEGER REFERENCES users(id),
    winner_id INTEGER REFERENCES users(id),
    status VARCHAR(20) DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Messages table
CREATE TABLE messages (
    id SERIAL PRIMARY KEY,
    room_id INTEGER REFERENCES rooms(id) ON DELETE CASCADE,
    user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    content TEXT NOT NULL,
    type VARCHAR(20) DEFAULT 'message',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX idx_users_rating ON users(rating DESC);
CREATE INDEX idx_rooms_status ON rooms(status);
CREATE INDEX idx_messages_room_created ON messages(room_id, created_at);
CREATE INDEX idx_tournaments_status ON tournaments(status);
CREATE INDEX idx_matches_tournament_round ON matches(tournament_id, round);

-- Insert default heroes for ZZZ
INSERT INTO heroes (name, element, rarity, role, description, image_url) VALUES
('Ellen Joe', 'Ice', 'S', 'Attack', 'Shark-themed ice attacker with high damage potential', ''),
('Zhu Yuan', 'Electric', 'S', 'Attack', 'Police officer with electric abilities and crowd control', ''),
('Jane Doe', 'Physical', 'S', 'Anomaly', 'Rat Thiren with physical anomaly buildup', ''),
('Lycaon', 'Ice', 'S', 'Stun', 'Wolf butler with ice stun abilities', ''),
('Koleda', 'Fire', 'S', 'Stun', 'President with fire stun and shield abilities', ''),
('Grace', 'Electric', 'S', 'Anomaly', 'Mechanic with electric anomaly and robot companion', ''),
('Rina', 'Electric', 'S', 'Support', 'Maid with electric support and pen ratio buffs', ''),
('Soldier 11', 'Fire', 'S', 'Attack', 'Military fire attacker with high burst damage', ''),
('Nekomata', 'Physical', 'S', 'Attack', 'Cat girl with dual blade physical attacks', ''),
('Anton', 'Electric', 'A', 'Attack', 'Bro with electric drill attacks', ''),
('Ben', 'Fire', 'A', 'Defense', 'Bear police officer with fire shield abilities', ''),
('Corin', 'Physical', 'A', 'Attack', 'Maid with physical chainsaw attacks', ''),
('Anby', 'Electric', 'A', 'Stun', 'Cunning Hares member with electric stun', ''),
('Nicole', 'Ether', 'A', 'Support', 'Cunning Hares leader with ether support', ''),
('Billy', 'Physical', 'A', 'Attack', 'Android cowboy with dual guns', ''),
('Soukaku', 'Ice', 'A', 'Support', 'Oni with ice support abilities', ''),
('Lucy', 'Fire', 'A', 'Support', 'Sons of Calydon member with fire support', ''),
('Piper', 'Physical', 'A', 'Anomaly', 'Sons of Calydon member with physical anomaly', '');
