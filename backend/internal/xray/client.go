package xray

type Client interface { AddUser(nodeID, userID string) error; RemoveUser(nodeID, userID string) error }