package logic

import ()

// Gets the events for a specific document.
type GetEventsInput struct {
	DocId string
	Db    DataCaller
}

type GetEventsOutput struct {
	Events []Event
}

func (lc Lgc) GetEvents(in *GetEventsInput) (*GetEventsOutput, error) {
	var DocEvents []Event

	// Retrieve all events for a document from the db.
	allEvents := []Event{}
	q := `SELECT id, body, created FROM events WHERE document_id = ? AND kind = ?;`
	err := in.Db.Select(&allEvents, q, in.DocId, "user")

	if err != nil {
		return &GetEventsOutput{DocEvents}, err
	}

	// Loop through each and write them to events for outputting.
	for _, event := range allEvents {
		e := Event{
			Id:      event.Id,
			Created: event.Created,
			Body:    event.Body,
		}
		DocEvents = append(DocEvents, e)
	}

	// Write the output and return.
	out := &GetEventsOutput{
		Events: DocEvents,
	}

	return out, err
}
