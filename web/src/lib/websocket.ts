import { authStorage } from './auth'

type EventHandler = (data: unknown) => void

class WebSocketManager {
  private ws: WebSocket | null = null
  private handlers: Map<string, EventHandler[]> = new Map()
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null
  private reconnectDelay = 3000
  private currentUrl: string | null = null

  connect(clusterId: string) {
    const token = authStorage.getAccessToken()
    if (!token) return

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const host = window.location.host
    const url = `${protocol}//${host}/ws/v1/clusters/${clusterId}/stream?token=${token}`
    this.currentUrl = url

    this.ws = new WebSocket(url)

    this.ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data)
        const eventType = data.type || 'message'
        const handlers = this.handlers.get(eventType) || []
        handlers.forEach((h) => h(data))
        const allHandlers = this.handlers.get('*') || []
        allHandlers.forEach((h) => h(data))
      } catch {
        // ignore parse errors
      }
    }

    this.ws.onclose = () => {
      if (this.currentUrl) {
        this.reconnectTimer = setTimeout(() => this.reconnect(), this.reconnectDelay)
      }
    }

    this.ws.onerror = () => {
      this.ws?.close()
    }
  }

  connectAll() {
    const token = authStorage.getAccessToken()
    if (!token) return

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const host = window.location.host
    const url = `${protocol}//${host}/ws/v1/clusters/stream?token=${token}`
    this.currentUrl = url

    this.ws = new WebSocket(url)

    this.ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data)
        const eventType = data.type || 'message'
        const handlers = this.handlers.get(eventType) || []
        handlers.forEach((h) => h(data))
        const allHandlers = this.handlers.get('*') || []
        allHandlers.forEach((h) => h(data))
      } catch {
        // ignore parse errors
      }
    }

    this.ws.onclose = () => {
      if (this.currentUrl) {
        this.reconnectTimer = setTimeout(() => this.reconnect(), this.reconnectDelay)
      }
    }

    this.ws.onerror = () => {
      this.ws?.close()
    }
  }

  private reconnect() {
    if (!this.currentUrl) return
    const token = authStorage.getAccessToken()
    if (!token) return

    const urlWithToken = this.currentUrl.replace(/\?token=.*$/, `?token=${token}`)
    this.ws = new WebSocket(urlWithToken)
    this.ws.onmessage = this.ws.onmessage
    this.ws.onclose = this.ws.onclose
    this.ws.onerror = this.ws.onerror
  }

  on(event: string, handler: EventHandler) {
    const handlers = this.handlers.get(event) || []
    handlers.push(handler)
    this.handlers.set(event, handlers)
  }

  off(event: string, handler: EventHandler) {
    const handlers = this.handlers.get(event) || []
    this.handlers.set(event, handlers.filter((h) => h !== handler))
  }

  disconnect() {
    this.currentUrl = null
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer)
      this.reconnectTimer = null
    }
    this.ws?.close()
    this.ws = null
    this.handlers.clear()
  }

  get isConnected() {
    return this.ws?.readyState === WebSocket.OPEN
  }
}

export const wsManager = new WebSocketManager()
