#!/bin/bash
set -e

move_if_needed() {
  src="$1"
  dst="$2"
  if [ "$src" != "$dst" ]; then
    mv "$src" "$dst"
  fi
}

# Projects
mkdir -p internal/projects
move_if_needed internal/projects/model.go internal/projects/model.go
move_if_needed internal/projects/db.go internal/projects/db.go
move_if_needed internal/projects/handlers.go internal/projects/handlers.go

# Quizzes
mkdir -p internal/quizzes
move_if_needed internal/quizzes/model.go internal/quizzes/model.go
move_if_needed internal/quizzes/db.go internal/quizzes/db.go
move_if_needed internal/quizzes/handlers.go internal/quizzes/handlers.go
move_if_needed internal/quizzes/results.go internal/quizzes/results.go
move_if_needed internal/quizzes/results_db.go internal/quizzes/results_db.go

# Revision
mkdir -p internal/revision
move_if_needed internal/revision/model.go internal/revision/model.go
move_if_needed internal/revision/db.go internal/revision/db.go
move_if_needed internal/revision/handlers.go internal/revision/handlers.go

# Resources
mkdir -p internal/resources
move_if_needed internal/resources/model.go internal/resources/model.go
move_if_needed internal/resources/db.go internal/resources/db.go
move_if_needed internal/resources/handlers.go internal/resources/handlers.go

# Leaderboard
mkdir -p internal/leaderboard
move_if_needed internal/leaderboard/model.go internal/leaderboard/model.go
move_if_needed internal/leaderboard/db.go internal/leaderboard/db.go
move_if_needed internal/leaderboard/handlers.go internal/leaderboard/handlers.go

# Achievements
mkdir -p internal/achievements
move_if_needed internal/achievements/model.go internal/achievements/model.go
move_if_needed internal/achievements/db.go internal/achievements/db.go
move_if_needed internal/achievements/handlers.go internal/achievements/handlers.go

# User
mkdir -p internal/user
move_if_needed internal/user/model.go internal/user/model.go
move_if_needed internal/user/db.go internal/user/db.go
move_if_needed internal/user/handlers.go internal/user/handlers.go

# Auth
mkdir -p internal/auth
move_if_needed internal/auth/jwt.go internal/auth/jwt.go

# Handlers (shared/middleware)
mkdir -p internal/handlers
move_if_needed internal/handlers/auth.go internal/handlers/auth.go
move_if_needed internal/handlers/error.go internal/handlers/error.go
move_if_needed internal/handlers/handlers.go internal/handlers/handlers.go
move_if_needed internal/handlers/health.go internal/handlers/health.go
move_if_needed internal/handlers/middleware.go internal/handlers/middleware.go
move_if_needed internal/handlers/pages.go internal/handlers/pages.go
move_if_needed internal/handlers/user.go internal/handlers/user.go

# Utils
mkdir -p internal/utils
move_if_needed internal/utils/publicid.go internal/utils/publicid.go
move_if_needed internal/utils/templui.go internal/utils/templui.go

# Remove empty dirs (if any)
find internal -type d -empty -delete

echo "Move complete. Project structure is now clean and idiomatic."
