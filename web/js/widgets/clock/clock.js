class ResilverClock extends HTMLElement {
  connectedCallback() {
    this.attachShadow({ mode: "open" });

    const cfg = JSON.parse(this.dataset.config || "{}");
    this._format = cfg.format || "24h";
    this._showSeconds = cfg.showSeconds !== false;
    this._showDate = cfg.showDate !== false;

    this.shadowRoot.innerHTML = `
      <style>
        :host {
          display: block;
          font-family: "SF Mono", "Fira Code", "Cascadia Code", monospace;
          color: #ddd;
          text-align: center;
        }
        .time {
          font-size: 4rem;
          font-weight: 300;
          letter-spacing: 0.05em;
        }
        .seconds {
          font-size: 2rem;
          opacity: 0.6;
        }
        .date {
          font-size: 1.2rem;
          opacity: 0.5;
          margin-top: 0.3em;
          font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
        }
      </style>
      <div class="time"></div>
      ${this._showDate ? '<div class="date"></div>' : ""}
    `;

    this._timeEl = this.shadowRoot.querySelector(".time");
    this._dateEl = this.shadowRoot.querySelector(".date");

    this._update();
    this._interval = setInterval(() => this._update(), 1000);
  }

  disconnectedCallback() {
    clearInterval(this._interval);
  }

  _update() {
    const now = new Date();
    let h = now.getHours();
    const m = String(now.getMinutes()).padStart(2, "0");
    const s = String(now.getSeconds()).padStart(2, "0");

    let suffix = "";
    if (this._format === "12h") {
      suffix = h >= 12 ? " PM" : " AM";
      h = h % 12 || 12;
    }

    const hStr = String(h).padStart(2, "0");
    let time = `${hStr}:${m}`;
    if (this._showSeconds) {
      time += `<span class="seconds">:${s}</span>`;
    }
    time += suffix;

    this._timeEl.innerHTML = time;

    if (this._dateEl) {
      this._dateEl.textContent = now.toLocaleDateString(undefined, {
        weekday: "long",
        year: "numeric",
        month: "long",
        day: "numeric",
      });
    }
  }
}

customElements.define("resilver-clock", ResilverClock);
