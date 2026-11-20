/**
 * A clean wrapper around EventSource that relies on the browser's
 * native reconnection and automatic Last-Event-ID header handling.
 */
export class SSESnk extends EventTarget {
  #url;
  #options;
  #eventSource = null;
  #customEvents = [];
  #sessionid=""
  #lastEventId = ""; // Track the latest ID across manual reconnects

  #watchdogTimer = null;
  #pingTimeoutMs = 35000; // Must be > server ping interval (20s)
  #retryCount = 0; //Track trety connecting during onerror

  constructor(url, options = {}) {
    super();
    this.#url = url;
    this.#options = options;
    this.#customEvents = options.customEvents || [];
  }

  session() {
    return this.#sessionid;
  }
  /**
   * Opens the SSE connection.
   */
  connect() {
    // Prevent duplicate connections if already open or actively retrying
    if (this.#eventSource) return;

    // Build URL and attach lastEventId as a query param if we have one
    const targetUrl = new URL(this.#url, window.location.origin);
    if (this.#lastEventId) {
      targetUrl.searchParams.set("lastEventId", this.#lastEventId);
    }

    this.#eventSource = new EventSource(targetUrl.toString(), {
      withCredentials: this.#options.withCredentials || false
    });

    this.#eventSource.addEventListener("sessionid", (event) => {
        console.log("sessionid",event.data)
    });

    // Connection opened (also fires whenever a native retry succeeds)
    this.#eventSource.onopen = (event) => {
      this.#resetWatchdog();
      this.#dispatchEvent('open', { url: this.#url });
    };

    // Generic unnamed messages
    this.#eventSource.onmessage = (event) => {
      this.#trackId(event);
      this.#resetWatchdog();
      this.#dispatchEvent('message', event);
    };

    // Named server heartbeat (event: ping)
    this.#eventSource.addEventListener("ping", (event) => {
      this.#trackId(event);
      this.#resetWatchdog();
      // Optional: forward "ping" if UI wants to show a connection indicator
      this.#dispatchEvent("ping", event);
    });

    // Native error & retry handling
    this.#eventSource.onerror = () => {
      // Guard in case disconnect() was called right as an error fired
      if (!this.#eventSource) return;

      const state = this.#eventSource.readyState;

      if (state === EventSource.CONNECTING) {
        this.#retryCount = (this.#retryCount || 0) + 1;

        // Keep the watchdog alive! If the browser's native retry hangs or stalls,
        // the watchdog will step in and force a clean reconnect with ?lastEventId=...
        this.#resetWatchdog();

        this.#dispatchEvent('reconnecting', {
          attempt: this.#retryCount,
          message: 'Connection lost. Browser retrying...',
        });
        return;
      }

      // EventSource.CLOSED (2): Fatal HTTP status (401, 403, 404, 500 with wrong Content-Type)
      // or browser aborted permanently.
      this.#clearWatchdog();
      this.#dispatchEvent('error', {
        fatal: true,
        readyState: state,
        message: 'SSE connection closed permanently by server or browser.',
      });
      this.disconnect();
    };
  }
  
  reconnect() {
    this.disconnect();
    this.connect();
  }

  #trackId(event) {
    if (event.lastEventId) {
      this.#parseeventid(event.lastEventId);
      this.#lastEventId = event.lastEventId;
    }
  }

  #parseeventid(eventid) {
    if ((eventid=eventid!==undefined && eventid!==null && typeof eventid==="string"?eventid:"")!=="") {
      if (eventid.lastIndexOf(":")>0){
        var newsessionid=eventid.substring(0,eventid.lastIndexOf(":"));
        if (newsessionid!=="" && this.#sessionid!==newsessionid) {
          this.#sessionid=newsessionid;
          this.#dispatchEvent("sessionid",newsessionid);
        }
      }
    }
    return eventid
  }

  #resetWatchdog() {
    this.#clearWatchdog();
    this.#watchdogTimer = setTimeout(() => {
      console.warn("SSE ping timeout: connection stale, reconnecting...");
      this.#dispatchEvent("stale", new CustomEvent("stale"));
      this.reconnect() // Force close and reconnect
    }, this.#pingTimeoutMs);
  }

  #clearWatchdog() {
    if (this.#watchdogTimer) {
      clearTimeout(this.#watchdogTimer);
      this.#watchdogTimer = null;
    }
  }

  /**
   * Helper to dispatch clean CustomEvents upstream.
   */
  #dispatchEvent(type, detail) {
    // 1. Safely parse lastEventId (fixed typo: lastEventId, and optional chaining in case detail is null)
    const lastEventId = this.#parseeventid(detail?.lastEventId) ?? "";
    
    // 2. Always create a NEW MessageEvent (never re-dispatch the original `detail` object)
    if (detail instanceof MessageEvent) {
      this.dispatchEvent(
        new MessageEvent(type, {
          data: detail.data,
          lastEventId: lastEventId,
          origin: detail.origin,
        })
      );
      return;
    }
    
    // 3. Always create a NEW Event for generic events like "open" or "error"
    if (detail instanceof Event) {
      this.dispatchEvent(new Event(type));
      return;
    }

    // 4. For plain custom payloads, attach lastEventId inside detail (or use MessageEvent)
    this.dispatchEvent(
      new MessageEvent(type, {
          data: detail,
          lastEventId: lastEventId,
        })
    );
  }

  /**
   * Explicitly closes the connection and stops native browser retries.
   */
  disconnect() {
    this.#clearWatchdog();
    if (this.#eventSource) {
      this.#eventSource.close();
      this.#eventSource = null;
      this.#dispatchEvent('close', { message: 'Connection explicitly closed.' });
    }
  }
}
