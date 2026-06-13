package service

import (
	"fmt"
	"sort"
	"strings"

	"devops-platform/internal/modules/notification/model"

	"gorm.io/gorm"
)

type NotificationService struct {
	db        *gorm.DB
	notifiers map[model.ChannelType]Notifier
	channels  []model.ChannelConfig
}

func NewNotificationService(db *gorm.DB) *NotificationService {
	return &NotificationService{
		db:        db,
		notifiers: make(map[model.ChannelType]Notifier),
	}
}

func (s *NotificationService) RegisterNotifier(channel model.ChannelType, notifier Notifier) {
	s.notifiers[channel] = notifier
}

func (s *NotificationService) Send(tenantID uint, channel model.ChannelType, recipients []string, subject, body string) error {
	notifier, ok := s.notifiers[channel]
	if !ok {
		return fmt.Errorf("no notifier registered for channel: %s", channel)
	}
	if err := notifier.Send(recipients, subject, body); err != nil {
		s.logSend(tenantID, channel, strings.Join(recipients, ","), body, false, err.Error())
		return err
	}
	s.logSend(tenantID, channel, strings.Join(recipients, ","), body, true, "")
	return nil
}

// SendWithFallback sends via the primary channel, falling back to lower priority channels on failure.
func (s *NotificationService) SendWithFallback(tenantID uint, recipients []string, subject, body string) error {
	var channels []struct {
		Channel  model.ChannelType
		Priority int
	}
	for ch, cfg := range s.channelPriority(tenantID) {
		channels = append(channels, struct {
			Channel  model.ChannelType
			Priority int
		}{ch, cfg})
	}
	sort.Slice(channels, func(i, j int) bool {
		return channels[i].Priority > channels[j].Priority
	})

	for _, ch := range channels {
		if err := s.Send(tenantID, ch.Channel, recipients, subject, body); err == nil {
			return nil
		}
	}
	return fmt.Errorf("all notification channels failed")
}

func (s *NotificationService) channelPriority(tenantID uint) map[model.ChannelType]int {
	var configs []model.ChannelConfig
	s.db.Where("tenant_id = ? AND enabled = ?", tenantID, true).Find(&configs)
	priorities := make(map[model.ChannelType]int)
	for _, c := range configs {
		priorities[c.Channel] = c.Priority
	}
	return priorities
}

func (s *NotificationService) logSend(tenantID uint, channel model.ChannelType, recipient, content string, success bool, errMsg string) {
	log := &model.SendLog{
		TenantID:  tenantID,
		Channel:   channel,
		Recipient: recipient,
		Content:   content,
		Success:   success,
		Error:     errMsg,
	}
	s.db.Create(log)
}
