package store

//go:generate go run -mod=mod entgo.io/ent/cmd/ent generate --feature sql/upsert --feature sql/modifier --feature sql/execquery ./schema --template ./templates/
