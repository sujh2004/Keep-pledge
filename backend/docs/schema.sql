CREATE DATABASE IF NOT EXISTS keep_pledge
  DEFAULT CHARACTER SET utf8mb4
  DEFAULT COLLATE utf8mb4_unicode_ci;

USE keep_pledge;

CREATE TABLE IF NOT EXISTS users (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  username VARCHAR(32) NOT NULL UNIQUE,
  email VARCHAR(128) NOT NULL UNIQUE,
  password_hash VARCHAR(255) NOT NULL,
  avatar VARCHAR(512) DEFAULT '',
  xp INT UNSIGNED DEFAULT 0,
  level INT UNSIGNED DEFAULT 1,
  credit_score INT DEFAULT 100,
  streak_freezes INT UNSIGNED DEFAULT 1,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS challenges (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  creator_id BIGINT UNSIGNED NOT NULL,
  title VARCHAR(128) NOT NULL,
  description TEXT,
  category ENUM('fitness','study','habit','health','social','other') DEFAULT 'other',
  pledge TEXT NOT NULL,
  penalty_type ENUM('donate','social','custom') NOT NULL,
  penalty_detail VARCHAR(512),
  challenge_type ENUM('daily','total') NOT NULL,
  target_days INT UNSIGNED NOT NULL,
  start_date DATE NOT NULL,
  end_date DATE NOT NULL,
  status ENUM('pending','active','completed','failed','cancelled') DEFAULT 'pending',
  is_public TINYINT(1) DEFAULT 0,
  xp_reward INT UNSIGNED DEFAULT 200,
  can_cancel_before DATETIME,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL,
  INDEX idx_challenges_creator (creator_id),
  INDEX idx_challenges_status (status),
  INDEX idx_challenges_public (is_public),
  CONSTRAINT fk_challenges_creator FOREIGN KEY (creator_id) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS challenge_participants (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  challenge_id BIGINT UNSIGNED NOT NULL,
  user_id BIGINT UNSIGNED NOT NULL,
  role ENUM('creator','participant','witness') NOT NULL,
  status ENUM('pending','accepted','rejected') DEFAULT 'pending',
  current_streak INT UNSIGNED DEFAULT 0,
  max_streak INT UNSIGNED DEFAULT 0,
  total_checkins INT UNSIGNED DEFAULT 0,
  joined_at DATETIME NOT NULL,
  UNIQUE KEY uk_challenge_user_role (challenge_id, user_id, role),
  INDEX idx_participants_user (user_id),
  CONSTRAINT fk_participants_challenge FOREIGN KEY (challenge_id) REFERENCES challenges(id),
  CONSTRAINT fk_participants_user FOREIGN KEY (user_id) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS checkins (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  challenge_id BIGINT UNSIGNED NOT NULL,
  user_id BIGINT UNSIGNED NOT NULL,
  content TEXT,
  image_url VARCHAR(512),
  checkin_date DATE NOT NULL,
  streak_count INT UNSIGNED DEFAULT 1,
  xp_earned INT UNSIGNED DEFAULT 10,
  like_count INT UNSIGNED DEFAULT 0,
  is_reported TINYINT(1) DEFAULT 0,
  created_at DATETIME NOT NULL,
  UNIQUE KEY uk_checkin_daily (challenge_id, user_id, checkin_date),
  INDEX idx_checkins_user (user_id),
  INDEX idx_checkins_date (checkin_date),
  CONSTRAINT fk_checkins_challenge FOREIGN KEY (challenge_id) REFERENCES challenges(id),
  CONSTRAINT fk_checkins_user FOREIGN KEY (user_id) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS checkin_interactions (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  checkin_id BIGINT UNSIGNED NOT NULL,
  user_id BIGINT UNSIGNED NOT NULL,
  type ENUM('like','comment') NOT NULL,
  content VARCHAR(512),
  created_at DATETIME NOT NULL,
  INDEX idx_interactions_checkin (checkin_id),
  INDEX idx_interactions_user (user_id),
  CONSTRAINT fk_interactions_checkin FOREIGN KEY (checkin_id) REFERENCES checkins(id),
  CONSTRAINT fk_interactions_user FOREIGN KEY (user_id) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS streak_freeze_logs (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  user_id BIGINT UNSIGNED NOT NULL,
  challenge_id BIGINT UNSIGNED NOT NULL,
  freeze_date DATE NOT NULL,
  created_at DATETIME NOT NULL,
  UNIQUE KEY uk_freeze_daily (user_id, challenge_id, freeze_date),
  CONSTRAINT fk_freezes_user FOREIGN KEY (user_id) REFERENCES users(id),
  CONSTRAINT fk_freezes_challenge FOREIGN KEY (challenge_id) REFERENCES challenges(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS achievements (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  name VARCHAR(64) NOT NULL UNIQUE,
  description VARCHAR(256) NOT NULL,
  icon VARCHAR(256),
  category ENUM('checkin','challenge','social','hidden') NOT NULL,
  condition_type VARCHAR(64) NOT NULL,
  condition_value INT NOT NULL,
  is_hidden TINYINT(1) DEFAULT 0,
  xp_reward INT UNSIGNED DEFAULT 50
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS user_achievements (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  user_id BIGINT UNSIGNED NOT NULL,
  achievement_id BIGINT UNSIGNED NOT NULL,
  unlocked_at DATETIME NOT NULL,
  UNIQUE KEY uk_user_achievement (user_id, achievement_id),
  CONSTRAINT fk_user_achievements_user FOREIGN KEY (user_id) REFERENCES users(id),
  CONSTRAINT fk_user_achievements_achievement FOREIGN KEY (achievement_id) REFERENCES achievements(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS friendships (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  user_id BIGINT UNSIGNED NOT NULL,
  friend_id BIGINT UNSIGNED NOT NULL,
  status ENUM('pending','accepted','blocked') DEFAULT 'pending',
  created_at DATETIME NOT NULL,
  UNIQUE KEY uk_friendship (user_id, friend_id),
  CONSTRAINT fk_friendships_user FOREIGN KEY (user_id) REFERENCES users(id),
  CONSTRAINT fk_friendships_friend FOREIGN KEY (friend_id) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS certificates (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  user_id BIGINT UNSIGNED NOT NULL,
  challenge_id BIGINT UNSIGNED NOT NULL,
  certificate_no VARCHAR(32) NOT NULL UNIQUE,
  image_url VARCHAR(512),
  issued_at DATETIME NOT NULL,
  CONSTRAINT fk_certificates_user FOREIGN KEY (user_id) REFERENCES users(id),
  CONSTRAINT fk_certificates_challenge FOREIGN KEY (challenge_id) REFERENCES challenges(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS notifications (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  user_id BIGINT UNSIGNED NOT NULL,
  type ENUM('invite','checkin_remind','streak_warn','achievement','witness','friend_request','like','comment','system') NOT NULL,
  title VARCHAR(128) NOT NULL,
  content VARCHAR(512) NOT NULL,
  related_id BIGINT UNSIGNED,
  related_type VARCHAR(32),
  is_read TINYINT(1) DEFAULT 0,
  created_at DATETIME NOT NULL,
  INDEX idx_notification_user_read (user_id, is_read),
  CONSTRAINT fk_notifications_user FOREIGN KEY (user_id) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS reports (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  reporter_id BIGINT UNSIGNED NOT NULL,
  target_type ENUM('checkin','challenge','user') NOT NULL,
  target_id BIGINT UNSIGNED NOT NULL,
  reason VARCHAR(512) NOT NULL,
  status ENUM('pending','resolved','dismissed') DEFAULT 'pending',
  created_at DATETIME NOT NULL,
  INDEX idx_reports_target (target_type, target_id),
  CONSTRAINT fk_reports_user FOREIGN KEY (reporter_id) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
