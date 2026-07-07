package reliable

// RetransmissionManager handles selective retransmission of unconfirmed packets.
type RetransmissionManager struct {
	checkpoint *Checkpoint
	sender     *Sender
}

// NewRetransmissionManager creates a manager linked to a checkpoint and sender.
func NewRetransmissionManager(checkpoint *Checkpoint, sender *Sender) *RetransmissionManager {
	return &RetransmissionManager{checkpoint: checkpoint, sender: sender}
}

// RetransmitUnconfirmed retransmits only unconfirmed packets from a failed hop.
// Confirmed packets remain confirmed; only the unconfirmed packet is resent.
func (rm *RetransmissionManager) RetransmitUnconfirmed(packets []*ReliablePacket, newNextPlanet string) []RetransmitResult {
	var results []RetransmitResult

	for _, pkt := range packets {
		if pkt.Status == StatusAcknowledged || pkt.Status == StatusDelivered {
			results = append(results, RetransmitResult{
				PacketID: pkt.PacketID,
				Action:   "KEEP_CONFIRMED",
				Success:  true,
			})
			continue
		}

		if pkt.Status == StatusInFlight || pkt.Status == StatusDeliveryUncertain {
			success := rm.sender.SendWithRetry(pkt, newNextPlanet)
			action := "RETRANSMIT"
			if !success {
				action = "RETRANSMIT_FAILED"
			}
			results = append(results, RetransmitResult{
				PacketID: pkt.PacketID,
				Action:   action,
				Success:  success,
			})
			continue
		}

		// Queued packets stay queued.
		results = append(results, RetransmitResult{
			PacketID: pkt.PacketID,
			Action:   "REMAIN_QUEUED",
			Success:  true,
		})
	}

	return results
}

// RetransmitResult describes the outcome of a retransmission attempt.
type RetransmitResult struct {
	PacketID string `json:"packet_id"`
	Action   string `json:"action"`
	Success  bool   `json:"success"`
}
