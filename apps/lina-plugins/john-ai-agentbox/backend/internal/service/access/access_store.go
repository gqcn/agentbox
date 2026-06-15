// This file implements DAO-backed AgentBox ownership predicates. Each query
// injects the AgentBox user boundary into the database condition so callers do
// not load cross-user resources before filtering.

package access

import (
	"context"
	"strings"

	"john-ai-agentbox/backend/internal/dao"
	"john-ai-agentbox/backend/internal/model/do"
)

type daoStore struct{}

var _ Store = (*daoStore)(nil)

// NewDAOStore creates the production access store backed by plugin-generated DAO.
func NewDAOStore() Store {
	return &daoStore{}
}

// UserOwnsAgent reports whether one non-deleted Agent belongs to userID.
func (s *daoStore) UserOwnsAgent(ctx context.Context, userID string, agentID string) (bool, error) {
	count, err := dao.CodingAgents.Ctx(ctx).
		Where(do.CodingAgents{
			Id:     strings.TrimSpace(agentID),
			UserId: strings.TrimSpace(userID),
		}).
		Count()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// UserOwnsChatSession reports whether sessionID belongs to agentID and userID.
func (s *daoStore) UserOwnsChatSession(ctx context.Context, userID string, agentID string, sessionID string) (bool, error) {
	count, err := dao.AgentChatSessions.Ctx(ctx).
		As("cs").
		InnerJoin(dao.CodingAgents.Table()+" a", "a.id = cs.agent_id").
		Where("a.user_id", strings.TrimSpace(userID)).
		Where("a.id", strings.TrimSpace(agentID)).
		Where("cs.id", strings.TrimSpace(sessionID)).
		Count()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// UserOwnsTerminalSession reports whether terminalID belongs to agentID and userID.
func (s *daoStore) UserOwnsTerminalSession(ctx context.Context, userID string, agentID string, terminalID string) (bool, error) {
	count, err := dao.AgentTerminalSessions.Ctx(ctx).
		Where(do.AgentTerminalSessions{
			UserId:     strings.TrimSpace(userID),
			AgentId:    strings.TrimSpace(agentID),
			TerminalId: strings.TrimSpace(terminalID),
		}).
		Count()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
