package aggregates_test

import (
	"strings"
	"testing"
	"time"

	"easi/backend/internal/archassistant/domain/aggregates"
	vo "easi/backend/internal/archassistant/domain/valueobjects"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	testTenantID      = uuid.NewString()
	testUserID        = uuid.NewString()
	testOtherUserID   = uuid.NewString()
	testConvID        = uuid.NewString()
	testMsgID1        = uuid.NewString()
)

func TestNewConversation(t *testing.T) {
	conv := aggregates.NewConversation(testTenantID, testUserID)

	assert.NotEmpty(t, conv.ID())
	assert.Equal(t, testTenantID, conv.TenantID())
	assert.Equal(t, testUserID, conv.UserID())
	assert.Equal(t, "New conversation", conv.Title())
	assert.False(t, conv.CreatedAt().IsZero())
	assert.False(t, conv.LastMessageAt().IsZero())
}

func TestConversation_AddUserMessage(t *testing.T) {
	conv := aggregates.NewConversation(testTenantID, testUserID)

	msg, err := conv.AddUserMessage("Hello, assistant!")
	require.NoError(t, err)

	assert.NotEmpty(t, msg.ID())
	assert.Equal(t, conv.ID(), msg.ConversationID())
	assert.Equal(t, vo.MessageRoleUser, msg.Role())
	assert.Equal(t, "Hello, assistant!", msg.Content())
	assert.Nil(t, msg.TokensUsed())
}

func TestConversation_AddUserMessage_GeneratesTitle(t *testing.T) {
	conv := aggregates.NewConversation(testTenantID, testUserID)
	assert.Equal(t, "New conversation", conv.Title())

	_, err := conv.AddUserMessage("What are our main capabilities?")
	require.NoError(t, err)

	assert.Equal(t, "What are our main capabilities?", conv.Title())
}

func TestConversation_AddUserMessage_TruncatesLongTitle(t *testing.T) {
	conv := aggregates.NewConversation(testTenantID, testUserID)

	longMsg := strings.Repeat("a", 200)
	_, err := conv.AddUserMessage(longMsg)
	require.NoError(t, err)

	assert.Len(t, conv.Title(), vo.MaxTitleLength)
	assert.True(t, strings.HasSuffix(conv.Title(), "..."))
}

func TestConversation_AddUserMessage_TitleSetOnlyOnce(t *testing.T) {
	conv := aggregates.NewConversation(testTenantID, testUserID)

	_, err := conv.AddUserMessage("First message")
	require.NoError(t, err)
	assert.Equal(t, "First message", conv.Title())

	_, err = conv.AddUserMessage("Second message")
	require.NoError(t, err)
	assert.Equal(t, "First message", conv.Title())
}

func TestConversation_AddUserMessage_EmptyContent(t *testing.T) {
	conv := aggregates.NewConversation(testTenantID, testUserID)
	_, err := conv.AddUserMessage("")
	assert.ErrorIs(t, err, vo.ErrMessageContentEmpty)
}

func TestConversation_AddUserMessage_TooLong(t *testing.T) {
	conv := aggregates.NewConversation(testTenantID, testUserID)
	_, err := conv.AddUserMessage(strings.Repeat("a", 2001))
	assert.ErrorIs(t, err, vo.ErrMessageContentTooLong)
}

func TestConversation_AddUserMessage_ExactlyMaxLength(t *testing.T) {
	conv := aggregates.NewConversation(testTenantID, testUserID)
	_, err := conv.AddUserMessage(strings.Repeat("a", 2000))
	assert.NoError(t, err)
}

func TestConversation_AddUserMessage_SanitizesControlChars(t *testing.T) {
	conv := aggregates.NewConversation(testTenantID, testUserID)
	msg, err := conv.AddUserMessage("hello\x00world\nkeep newlines")
	require.NoError(t, err)
	assert.Equal(t, "helloworld\nkeep newlines", msg.Content())
}

func TestConversation_AddUserMessage_TitleSanitizesAngleBrackets(t *testing.T) {
	conv := aggregates.NewConversation(testTenantID, testUserID)
	_, err := conv.AddUserMessage("<script>alert('xss')</script>")
	require.NoError(t, err)
	assert.NotContains(t, conv.Title(), "<")
	assert.NotContains(t, conv.Title(), ">")
}

func TestConversation_AddAssistantMessage(t *testing.T) {
	conv := aggregates.NewConversation(testTenantID, testUserID)

	msg := conv.AddAssistantMessage("Here is your answer.", 42)

	assert.NotEmpty(t, msg.ID())
	assert.Equal(t, conv.ID(), msg.ConversationID())
	assert.Equal(t, vo.MessageRoleAssistant, msg.Role())
	assert.Equal(t, "Here is your answer.", msg.Content())
	require.NotNil(t, msg.TokensUsed())
	assert.Equal(t, 42, *msg.TokensUsed())
}

func TestConversation_IsOwnedBy(t *testing.T) {
	conv := aggregates.NewConversation(testTenantID, testUserID)
	assert.True(t, conv.IsOwnedBy(testUserID))
	assert.False(t, conv.IsOwnedBy(testOtherUserID))
}

func newConvWithStaleLastMessageAt() (*aggregates.Conversation, time.Time) {
	conv := aggregates.ReconstructConversation(aggregates.ReconstructConversationParams{
		ID:            testConvID,
		TenantID:      testTenantID,
		UserID:        testUserID,
		Title:         "Existing title",
		LastMessageAt: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
	})
	return conv, time.Now().Add(-time.Millisecond)
}

func TestConversation_AddUserMessage_UpdatesLastMessageAt(t *testing.T) {
	conv, before := newConvWithStaleLastMessageAt()

	_, err := conv.AddUserMessage("Hello")
	require.NoError(t, err)

	assert.True(t, conv.LastMessageAt().After(before), "lastMessageAt should be updated to current time")
}

func TestConversation_AddAssistantMessage_UpdatesLastMessageAt(t *testing.T) {
	conv, before := newConvWithStaleLastMessageAt()

	conv.AddAssistantMessage("Response", 42)

	assert.True(t, conv.LastMessageAt().After(before), "lastMessageAt should be updated to current time")
}

func TestMessageRole_IsValid(t *testing.T) {
	assert.True(t, vo.MessageRoleUser.IsValid())
	assert.True(t, vo.MessageRoleAssistant.IsValid())
	assert.True(t, vo.MessageRoleTool.IsValid())
	assert.False(t, vo.MessageRole("invalid").IsValid())
}

func TestNewMessageRole(t *testing.T) {
	role, err := vo.NewMessageRole("user")
	require.NoError(t, err)
	assert.Equal(t, vo.MessageRoleUser, role)

	_, err = vo.NewMessageRole("invalid")
	assert.ErrorIs(t, err, vo.ErrInvalidMessageRole)
}

func TestReconstructConversation(t *testing.T) {
	conv := aggregates.ReconstructConversation(aggregates.ReconstructConversationParams{
		ID:       testConvID,
		TenantID: testTenantID,
		UserID:   testUserID,
		Title:    "Test conversation",
	})

	assert.Equal(t, testConvID, conv.ID())
	assert.Equal(t, testTenantID, conv.TenantID())
	assert.Equal(t, testUserID, conv.UserID())
	assert.Equal(t, "Test conversation", conv.Title())
}

func TestReconstructMessage(t *testing.T) {
	tokens := 50
	msg := aggregates.ReconstructMessage(aggregates.ReconstructMessageParams{
		ID:             testMsgID1,
		ConversationID: testConvID,
		Role:           vo.MessageRoleAssistant,
		Content:        "Hello",
		TokensUsed:     &tokens,
	})

	assert.Equal(t, testMsgID1, msg.ID())
	assert.Equal(t, testConvID, msg.ConversationID())
	assert.Equal(t, vo.MessageRoleAssistant, msg.Role())
	assert.Equal(t, "Hello", msg.Content())
	require.NotNil(t, msg.TokensUsed())
	assert.Equal(t, 50, *msg.TokensUsed())
}
