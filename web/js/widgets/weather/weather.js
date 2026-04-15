class ResilverWeather extends HTMLElement {
  connectedCallback() {
    const cfg = JSON.parse(this.dataset.config || "{}");
    this._units = cfg.units || "celsius";
    this._location = cfg.location || "";
    this._refreshInterval = (cfg.refreshIntervalSeconds || 1800) * 1000;

    this.className = "flex flex-col justify-center items-center w-full h-full text-gray-300 text-center";
    this.innerHTML = `
      ${this._location ? `<div class="resilver-weather__location opacity-50 mb-1" style="font-size: 1.5cqmin">${this._location}</div>` : ""}
      <div class="resilver-weather__icon" style="font-size: 10cqmin"></div>
      <div class="resilver-weather__temp font-light mt-1" style="font-size: 7cqmin; color: var(--accent, inherit)"></div>
      <div class="resilver-weather__desc opacity-60 mt-0.5" style="font-size: 2.5cqmin"></div>
      <div class="resilver-weather__details opacity-40 mt-1" style="font-size: 2cqmin"></div>
    `;

    this._iconEl = this.querySelector(".resilver-weather__icon");
    this._tempEl = this.querySelector(".resilver-weather__temp");
    this._descEl = this.querySelector(".resilver-weather__desc");
    this._detailsEl = this.querySelector(".resilver-weather__details");

    this._fetch();
    this._interval = setInterval(() => this._fetch(), this._refreshInterval);
  }

  disconnectedCallback() {
    clearInterval(this._interval);
  }

  async _fetch() {
    try {
      const resp = await fetch("/api/weather");
      if (!resp.ok) throw new Error(`HTTP ${resp.status}`);
      const data = await resp.json();
      this._render(data);
    } catch (err) {
      this._iconEl.textContent = "";
      this._tempEl.textContent = "--";
      this._descEl.textContent = "Unable to load weather";
      this._detailsEl.textContent = "";
    }
  }

  _render(data) {
    const unit = this._units === "fahrenheit" ? "°F" : "°C";

    this._iconEl.innerHTML = `<i class="wi wi-${data.icon}"></i>`;
    this._tempEl.textContent = `${Math.round(data.temperature)}${unit}`;
    this._descEl.textContent = data.description;

    this._detailsEl.textContent = `Feels ${Math.round(data.apparentTemperature)}${unit}`;
  }
}

customElements.define("resilver-weather", ResilverWeather);
