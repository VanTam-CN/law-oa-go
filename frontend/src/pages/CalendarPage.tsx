import React, { useState } from 'react';
import { Calendar, momentLocalizer } from 'react-big-calendar';
import moment from 'moment';
import 'react-big-calendar/lib/css/react-big-calendar.css';
import {
  Badge,
  Dropdown,
  Button,
  Card,
  Modal,
  Form,
  Row,
  Col
} from 'react-bootstrap';

// 设置本地化
const localizer = momentLocalizer(moment);

// 事件类型定义
interface CalendarEvent {
  id: number;
  title: string;
  start: Date;
  end: Date;
  allDay?: boolean;
  resource?: any;
  type: 'hearing' | 'meeting' | 'deadline' | 'conference';
  priority: 'low' | 'medium' | 'high' | 'urgent';
  status: 'scheduled' | 'completed' | 'cancelled';
  description?: string;
  attendees?: string[];
}

const CalendarPage: React.FC = () => {
  const [events, setEvents] = useState<CalendarEvent[]>([
    {
      id: 1,
      title: 'Case Hearing #1234',
      start: new Date(2025, 8, 15, 10, 0),
      end: new Date(2025, 8, 15, 12, 0),
      type: 'hearing',
      priority: 'high',
      status: 'scheduled',
      description: 'Civil case hearing for John Doe vs. ABC Corporation',
      attendees: ['Jane Smith', 'John Doe']
    },
    {
      id: 2,
      title: 'Client Meeting',
      start: new Date(2025, 8, 16, 14, 0),
      end: new Date(2025, 8, 16, 15, 30),
      type: 'meeting',
      priority: 'medium',
      status: 'scheduled',
      description: 'Meeting with client to discuss case progress',
      attendees: ['Michael Johnson', 'Sarah Williams']
    },
    {
      id: 3,
      title: 'Document Submission Deadline',
      start: new Date(2025, 8, 18, 17, 0),
      end: new Date(2025, 8, 18, 17, 0),
      allDay: true,
      type: 'deadline',
      priority: 'urgent',
      status: 'scheduled',
      description: 'Deadline for submitting case documents for Case #5678'
    },
    {
      id: 4,
      title: 'Legal Conference',
      start: new Date(2025, 8, 20, 9, 0),
      end: new Date(2025, 8, 20, 17, 0),
      type: 'conference',
      priority: 'medium',
      status: 'scheduled',
      description: 'Annual legal conference at Grand Hotel',
      attendees: ['Entire Legal Team']
    }
  ]);

  const [showEventModal, setShowEventModal] = useState(false);
  const [selectedEvent, setSelectedEvent] = useState<CalendarEvent | null>(null);
  const [view, setView] = useState('month');
  const [date, setDate] = useState(new Date());

  const handleSelectEvent = (event: CalendarEvent) => {
    setSelectedEvent(event);
    setShowEventModal(true);
  };

  const handleSelectSlot = (slotInfo: { start: Date; end: Date; slots: Date[] }) => {
    const newEvent: CalendarEvent = {
      id: events.length + 1,
      title: 'New Event',
      start: slotInfo.start,
      end: slotInfo.end,
      type: 'meeting',
      priority: 'medium',
      status: 'scheduled'
    };
    setSelectedEvent(newEvent);
    setShowEventModal(true);
  };

  const handleCloseEventModal = () => {
    setShowEventModal(false);
    setSelectedEvent(null);
  };

  const eventStyleGetter = (event: CalendarEvent) => {
    let backgroundColor = '';
    switch (event.type) {
      case 'hearing':
        backgroundColor = '#dc3545'; // Red
        break;
      case 'meeting':
        backgroundColor = '#007bff'; // Blue
        break;
      case 'deadline':
        backgroundColor = '#ffc107'; // Yellow
        break;
      case 'conference':
        backgroundColor = '#28a745'; // Green
        break;
      default:
        backgroundColor = '#6c757d'; // Gray
    }

    const style = {
      backgroundColor,
      borderRadius: '5px',
      opacity: 0.8,
      color: 'white',
      border: '0px',
      display: 'block'
    };

    return {
      style
    };
  };

  const EventTypeBadge = ({ type }: { type: string }) => {
    let badgeClass = '';
    let iconClass = '';
    let text = '';

    switch (type) {
      case 'hearing':
        badgeClass = 'bg-danger';
        iconClass = 'fas fa-gavel';
        text = 'Hearing';
        break;
      case 'meeting':
        badgeClass = 'bg-primary';
        iconClass = 'fas fa-users';
        text = 'Meeting';
        break;
      case 'deadline':
        badgeClass = 'bg-warning';
        iconClass = 'fas fa-clock';
        text = 'Deadline';
        break;
      case 'conference':
        badgeClass = 'bg-success';
        iconClass = 'fas fa-building';
        text = 'Conference';
        break;
      default:
        badgeClass = 'bg-secondary';
        iconClass = 'fas fa-calendar';
        text = type;
    }

    return (
      <Badge bg="" className={`${badgeClass} me-2`}>
        <i className={`${iconClass} me-1`}></i>
        {text}
      </Badge>
    );
  };

  const PriorityBadge = ({ priority }: { priority: string }) => {
    let badgeClass = '';
    let text = '';

    switch (priority) {
      case 'low':
        badgeClass = 'bg-info';
        text = 'Low';
        break;
      case 'medium':
        badgeClass = 'bg-warning';
        text = 'Medium';
        break;
      case 'high':
        badgeClass = 'bg-danger';
        text = 'High';
        break;
      case 'urgent':
        badgeClass = 'bg-danger';
        text = 'Urgent';
        break;
      default:
        badgeClass = 'bg-secondary';
        text = priority;
    }

    return (
      <Badge bg="" className={badgeClass}>
        {text}
      </Badge>
    );
  };

  const StatusBadge = ({ status }: { status: string }) => {
    let badgeClass = '';
    let text = '';

    switch (status) {
      case 'scheduled':
        badgeClass = 'bg-primary';
        text = 'Scheduled';
        break;
      case 'completed':
        badgeClass = 'bg-success';
        text = 'Completed';
        break;
      case 'cancelled':
        badgeClass = 'bg-secondary';
        text = 'Cancelled';
        break;
      default:
        badgeClass = 'bg-secondary';
        text = status;
    }

    return (
      <Badge bg="" className={badgeClass}>
        {text}
      </Badge>
    );
  };

  return (
    <div>
      <div className="d-flex justify-content-between align-items-center mb-4">
        <h1>Calendar</h1>
        <div className="d-flex">
          <Dropdown className="me-2">
            <Dropdown.Toggle variant="outline-secondary" id="view-dropdown">
              <i className="fas fa-eye me-2"></i>
              View: {view.charAt(0).toUpperCase() + view.slice(1)}
            </Dropdown.Toggle>
            <Dropdown.Menu>
              <Dropdown.Item onClick={() => setView('month')}>Month</Dropdown.Item>
              <Dropdown.Item onClick={() => setView('week')}>Week</Dropdown.Item>
              <Dropdown.Item onClick={() => setView('day')}>Day</Dropdown.Item>
              <Dropdown.Item onClick={() => setView('agenda')}>Agenda</Dropdown.Item>
            </Dropdown.Menu>
          </Dropdown>
          <Button variant="primary" onClick={() => setShowEventModal(true)}>
            <i className="fas fa-plus me-2"></i>
            Add Event
          </Button>
        </div>
      </div>

      <Card className="mb-4">
        <Card.Body>
          <Calendar
            localizer={localizer}
            events={events}
            startAccessor="start"
            endAccessor="end"
            style={{ height: 700 }}
            onView={(view) => setView(view)}
            onNavigate={(date) => setDate(date)}
            view={view as any}
            date={date}
            onSelectEvent={handleSelectEvent}
            onSelectSlot={handleSelectSlot}
            selectable
            eventPropGetter={eventStyleGetter}
            views={['month', 'week', 'day', 'agenda']}
            messages={{
              next: 'Next',
              previous: 'Back',
              today: 'Today',
              month: 'Month',
              week: 'Week',
              day: 'Day',
              agenda: 'Agenda'
            }}
          />
        </Card.Body>
      </Card>

      {/* Event Modal */}
      <Modal show={showEventModal} onHide={handleCloseEventModal} size="lg">
        <Modal.Header closeButton>
          <Modal.Title>
            {selectedEvent ? (
              <span><i className="fas fa-calendar-plus me-2"></i> Edit Event</span>
            ) : (
              <span><i className="fas fa-calendar-plus me-2"></i> Add Event</span>
            )}
          </Modal.Title>
        </Modal.Header>
        <Form>
          <Modal.Body>
            {selectedEvent && (
              <>
                <Row>
                  <Col md={12}>
                    <Form.Group className="mb-3">
                      <Form.Label>Title <span className="text-danger">*</span></Form.Label>
                      <Form.Control
                        type="text"
                        value={selectedEvent.title}
                        onChange={(e) => setSelectedEvent({...selectedEvent, title: e.target.value})}
                        required
                        placeholder="Enter event title"
                      />
                    </Form.Group>
                  </Col>
                </Row>
                <Row>
                  <Col md={6}>
                    <Form.Group className="mb-3">
                      <Form.Label>Start Date & Time <span className="text-danger">*</span></Form.Label>
                      <Form.Control
                        type="datetime-local"
                        value={moment(selectedEvent.start).format('YYYY-MM-DDTHH:mm')}
                        onChange={(e) => setSelectedEvent({...selectedEvent, start: new Date(e.target.value)})}
                        required
                      />
                    </Form.Group>
                  </Col>
                  <Col md={6}>
                    <Form.Group className="mb-3">
                      <Form.Label>End Date & Time <span className="text-danger">*</span></Form.Label>
                      <Form.Control
                        type="datetime-local"
                        value={moment(selectedEvent.end).format('YYYY-MM-DDTHH:mm')}
                        onChange={(e) => setSelectedEvent({...selectedEvent, end: new Date(e.target.value)})}
                        required
                      />
                    </Form.Group>
                  </Col>
                </Row>
                <Row>
                  <Col md={4}>
                    <Form.Group className="mb-3">
                      <Form.Label>Type</Form.Label>
                      <Form.Select
                        value={selectedEvent.type}
                        onChange={(e) => setSelectedEvent({...selectedEvent, type: e.target.value as any})}
                      >
                        <option value="hearing">Hearing</option>
                        <option value="meeting">Meeting</option>
                        <option value="deadline">Deadline</option>
                        <option value="conference">Conference</option>
                      </Form.Select>
                    </Form.Group>
                  </Col>
                  <Col md={4}>
                    <Form.Group className="mb-3">
                      <Form.Label>Priority</Form.Label>
                      <Form.Select
                        value={selectedEvent.priority}
                        onChange={(e) => setSelectedEvent({...selectedEvent, priority: e.target.value as any})}
                      >
                        <option value="low">Low</option>
                        <option value="medium">Medium</option>
                        <option value="high">High</option>
                        <option value="urgent">Urgent</option>
                      </Form.Select>
                    </Form.Group>
                  </Col>
                  <Col md={4}>
                    <Form.Group className="mb-3">
                      <Form.Label>Status</Form.Label>
                      <Form.Select
                        value={selectedEvent.status}
                        onChange={(e) => setSelectedEvent({...selectedEvent, status: e.target.value as any})}
                      >
                        <option value="scheduled">Scheduled</option>
                        <option value="completed">Completed</option>
                        <option value="cancelled">Cancelled</option>
                      </Form.Select>
                    </Form.Group>
                  </Col>
                </Row>
                <Form.Group className="mb-3">
                  <Form.Label>Description</Form.Label>
                  <Form.Control
                    as="textarea"
                    rows={3}
                    value={selectedEvent.description || ''}
                    onChange={(e) => setSelectedEvent({...selectedEvent, description: e.target.value})}
                    placeholder="Enter event description"
                  />
                </Form.Group>
                <Form.Group className="mb-3">
                  <Form.Label>Attendees</Form.Label>
                  <Form.Control
                    type="text"
                    value={selectedEvent.attendees?.join(', ') || ''}
                    onChange={(e) => setSelectedEvent({...selectedEvent, attendees: e.target.value.split(', ')})}
                    placeholder="Enter attendee names (comma separated)"
                  />
                </Form.Group>
              </>
            )}
          </Modal.Body>
          <Modal.Footer>
            <Button variant="secondary" onClick={handleCloseEventModal}>
              <i className="fas fa-times me-2"></i>
              Cancel
            </Button>
            <Button variant="primary">
              {selectedEvent ? (
                <span><i className="fas fa-save me-2"></i> Update Event</span>
              ) : (
                <span><i className="fas fa-plus me-2"></i> Add Event</span>
              )}
            </Button>
          </Modal.Footer>
        </Form>
      </Modal>

      {/* Upcoming Events */}
      <Card>
        <Card.Header className="d-flex justify-content-between align-items-center">
          <span><i className="fas fa-calendar-check me-2"></i> Upcoming Events</span>
          <Button variant="outline-primary" size="sm">
            <i className="fas fa-external-link-alt me-2"></i>
            View All
          </Button>
        </Card.Header>
        <Card.Body>
          <div className="upcoming-events">
            {events
              .filter(event => event.start > new Date())
              .sort((a, b) => a.start.getTime() - b.start.getTime())
              .slice(0, 5)
              .map(event => (
                <div key={event.id} className="event-item mb-3 p-3 border rounded">
                  <div className="d-flex justify-content-between align-items-center mb-2">
                    <h6 className="mb-0">{event.title}</h6>
                    <div>
                      <EventTypeBadge type={event.type} />
                      <PriorityBadge priority={event.priority} />
                    </div>
                  </div>
                  <p className="mb-2 text-muted">{event.description}</p>
                  <div className="d-flex justify-content-between align-items-center">
                    <small className="text-muted">
                      <i className="fas fa-calendar me-1"></i>
                      {moment(event.start).format('MMM DD, YYYY HH:mm')}
                    </small>
                    <small className="text-muted">
                      <i className="fas fa-users me-1"></i>
                      {event.attendees?.length || 0} attendees
                    </small>
                  </div>
                </div>
              ))
            }
            
            {events.filter(event => event.start > new Date()).length === 0 && (
              <div className="text-center py-3">
                <i className="fas fa-calendar-plus fa-2x text-muted mb-3"></i>
                <h6>No upcoming events</h6>
                <p className="text-muted">Add events to your calendar to stay organized</p>
                <Button variant="primary" onClick={() => setShowEventModal(true)}>
                  <i className="fas fa-plus me-2"></i>
                  Add Your First Event
                </Button>
              </div>
            )}
          </div>
        </Card.Body>
      </Card>
    </div>
  );
};

export default CalendarPage;