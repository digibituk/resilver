class ResilverClock extends HTMLElement {
  connectedCallback() {
    const cfg = JSON.parse(this.dataset.config || "{}");
    this._format = cfg.format || "24h";
    this._showSeconds = cfg.showSeconds !== false;
    this._showDate = cfg.showDate !== false;
    this._timezone = cfg.timezone || undefined;

    this.className = "flex flex-col justify-center items-center w-full h-full font-mono text-gray-300 text-center";
    this.innerHTML = `
      <div class="resilver-clock__time font-light tracking-wider" style="font-size: 14cqmin"></div>
      ${this._showDate ? '<div class="resilver-clock__date opacity-50 mt-1 font-sans" style="font-size: 3cqmin"></div>' : ""}
    `;

    this._timeEl = this.querySelector(".resilver-clock__time");
    this._dateEl = this.querySelector(".resilver-clock__date");

    this._update();
    this._interval = setInterval(() => this._update(), 1000);
  }

  disconnectedCallback() {
    clearInterval(this._interval);
  }

  _update() {
    const now = new Date();
    const timeOpts = {
      hour: "2-digit",
      minute: "2-digit",
      hour12: this._format === "12h",
    };
    if (this._showSeconds) {
      timeOpts.second = "2-digit";
    }
    if (this._timezone) {
      timeOpts.timeZone = this._timezone;
    }

    this._timeEl.textContent = now.toLocaleTimeString(undefined, timeOpts);

    if (this._dateEl) {
      const dateOpts = {
        weekday: "long",
        year: "numeric",
        month: "long",
        day: "numeric",
      };
      if (this._timezone) {
        dateOpts.timeZone = this._timezone;
      }
      this._dateEl.textContent = now.toLocaleDateString(undefined, dateOpts);
    }
  }
}

customElements.define("resilver-clock", ResilverClock);
