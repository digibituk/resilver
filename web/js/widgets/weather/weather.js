class ResilverWeather extends HTMLElement {
  connectedCallback() {
    const cfg = JSON.parse(this.dataset.config || "{}");
    this._units = cfg.units || "celsius";
    this._location = cfg.location || "";
    this._refreshInterval = (cfg.refreshIntervalSeconds || 1800) * 1000;

    this.className = "block text-gray-300 text-center";
    this.innerHTML = `
      ${this._location ? `<div class="resilver-weather__location text-sm opacity-50 mb-1">${this._location}</div>` : ""}
      <div class="resilver-weather__icon text-5xl"></div>
      <div class="resilver-weather__temp text-3xl font-light mt-1"></div>
      <div class="resilver-weather__desc text-sm opacity-60 mt-0.5"></div>
      <div class="resilver-weather__details text-xs opacity-40 mt-1"></div>
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
      this._iconEl.textContent = "⚠️";
      this._tempEl.textContent = "--";
      this._descEl.textContent = "Unable to load weather";
      this._detailsEl.textContent = "";
    }
  }

  _render(data) {
    const unit = this._units === "fahrenheit" ? "°F" : "°C";

    this._iconEl.textContent = data.icon;
    this._tempEl.textContent = `${Math.round(data.temperature)}${unit}`;
    this._descEl.textContent = data.description;

    const feelsLike = `Feels ${Math.round(data.apparentTemperature)}${unit}`;
    const humidity = `Humidity ${data.humidity}%`;
    this._detailsEl.textContent = `${feelsLike} · ${humidity}`;
  }
}

customElements.define("resilver-weather", ResilverWeather);
