export interface EventData {
  NodeName: string,
  CheckId: string,
  CheckDescription: string,
  StatusHistory: number,
  Status: number,
  Message: string,
  Timestamp: string,
}

export class EventStream {
  static EVENTS_ENDPOINT = '/api/events';
  static SOCKET_ENDPOINT = '/api/events/socket';

  static get SOCKET_URL(): string {
    const url = new URL(window.location.toString());
    url.pathname = EventStream.SOCKET_ENDPOINT;
    url.protocol = (window.location.protocol === 'https:') ? 'wss:' : 'ws:';
    return url.toString();
  }

  private ws: WebSocket | null;
  private heartbeat: number | null;
  private handler: (event: EventData) => void;

  constructor(handler: (event: EventData) => void) {
    this.ws = null;
    this.heartbeat = null;
    this.handler = handler;
  }

  connect(): void {
    if (this.isConnecting || this.isOpen) return;

    this.ws = new WebSocket(EventStream.SOCKET_URL);
    this.ws.addEventListener('close', () => this.reconnect());
    this.ws.addEventListener('error', () => this.reconnect());
    this.ws.addEventListener('message', (e) => this.handleMessage(e.data));

    if (this.heartbeat === null) {
      this.heartbeat = setInterval(() => this.sendHeartbeat(), 25_000);
    }
    this.fetchEvents();
  }

  private sendHeartbeat() {
    if (this.isOpen) this.ws?.send('💓');
  }

  private async fetchEvents() {
    const response = await fetch(EventStream.EVENTS_ENDPOINT);
    const events = await response.json();
    events.forEach((event: EventData) => this.handler(event));
  }

  private reconnect(): void {
    this.ws?.close();
    setTimeout(() => this.connect(), 1_000);
  }

  private get isConnecting(): boolean {
    return this.ws?.readyState === WebSocket.CONNECTING;
  }

  private get isOpen(): boolean {
    return this.ws?.readyState === WebSocket.OPEN;
  }

  private handleMessage(data: string): void {
    const event = JSON.parse(data);
    this.handler(event);
  }
}
