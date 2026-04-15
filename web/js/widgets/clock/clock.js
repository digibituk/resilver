class ResilverClock extends HTMLElement {
  connectedCallback() {
    const cfg = JSON.parse(this.dataset.config || "{}");
    this._format = cfg.format || "24h";
    this._showSeconds = cfg.showSeconds !== false;
    this._showDate = cfg.showDate !== false;
    this._timezone = cfg.timezone || undefined;

    this.className = "flex flex-col justify-center items-center w-full h-full text-gray-300 text-center";
    this.innerHTML = `
      <div class="resilver-clock__time text-[14cqmin] tabular-nums font-light tracking-wider">
        <span class="resilver-clock__hm"></span>${this._showSeconds ? '<span class="resilver-clock__sec text-[0.45em] opacity-40"></span>' : ""}
      </div>
      ${this._showDate ? '<div class="resilver-clock__date text-[4.5cqmin] accent opacity-70 mt-1"></div>' : ""}
    `;

    this._hmEl = this.querySelector(".resilver-clock__hm");
    this._secEl = this.querySelector(".resilver-clock__sec");
    this._dateEl = this.querySelector(".resilver-clock__date");

    this._update();
    this._interval = setInterval(() => this._update(), 1000);
  }

  disconnectedCallback() {
    clearInterval(this._interval);
  }

  _update() {
    const now = new Date();
    const localeOpts = { timeZone: this._timezone };
    const h12 = this._format === "12h";

    const hm = now.toLocaleTimeString(undefined, { ...localeOpts, hour: "2-digit", minute: "2-digit", hour12: h12 });

    this._hmEl.textContent = hm;

    if (this._secEl) {
      const s = now.toLocaleTimeString(undefined, { ...localeOpts, second: "2-digit", hour12: false }).slice(-2);
      this._secEl.style.opacity = "0.15";
      this._secEl.textContent = s;
      requestAnimationFrame(() => {
        this._secEl.style.opacity = "";
      });
    }

    if (this._dateEl) {
      this._dateEl.textContent = now.toLocaleDateString(undefined, {
        ...localeOpts,
        weekday: "long",
        year: "numeric",
        month: "long",
        day: "numeric",
      });
    }
  }
}

customElements.define("resilver-clock", ResilverClock);
