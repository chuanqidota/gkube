package auth

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

// ticketTTL ticket 有效期,WebSocket 一次性短票据。
const ticketTTL = 30 * time.Second

type ticketEntry struct {
	userID    uint
	username  string
	expiresAt time.Time
}

var (
	tickets   = make(map[string]ticketEntry)
	ticketsMu sync.Mutex
)

// IssueTicket 为指定用户签发一个 30 秒有效的一次性 ticket。
func IssueTicket(userID uint, username string) string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		b = []byte(time.Now().String() + username)
	}
	ticket := hex.EncodeToString(b)

	ticketsMu.Lock()
	tickets[ticket] = ticketEntry{
		userID:    userID,
		username:  username,
		expiresAt: time.Now().Add(ticketTTL),
	}
	// 顺带清理过期项(懒清理)
	now := time.Now()
	for k, v := range tickets {
		if !now.Before(v.expiresAt) {
			delete(tickets, k)
		}
	}
	ticketsMu.Unlock()

	return ticket
}

// ConsumeTicket 校验并消费 ticket,成功后立即删除(一次性),返回 userID 与 username。
// 失败或过期返回 ok=false。
func ConsumeTicket(ticket string) (userID uint, username string, ok bool) {
	if ticket == "" {
		return 0, "", false
	}
	ticketsMu.Lock()
	defer ticketsMu.Unlock()
	entry, exists := tickets[ticket]
	if !exists {
		return 0, "", false
	}
	delete(tickets, ticket)
	if !time.Now().Before(entry.expiresAt) {
		return 0, "", false
	}
	return entry.userID, entry.username, true
}
