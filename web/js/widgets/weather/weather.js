class ResilverWeather extends HTMLElement {
  connectedCallback() {
    const cfg = JSON.parse(this.dataset.config || "{}");
    this._units = cfg.units || "celsius";
    this._location = cfg.location || "";
    this._refreshInterval = (cfg.refreshIntervalSeconds || 1800) * 1000;

    this.className =
      "flex flex-col justify-center items-center w-full h-full text-gray-300 text-center";
    this.innerHTML = `
      <div class="flex items-center gap-[2.5cqmin]">
        <div class="resilver-weather__icon text-[10cqmin] animate-breathe"></div>
        <div class="resilver-weather__temp text-[10cqmin] font-light"></div>
      </div>
      <div class="resilver-weather__desc text-[3cqmin] opacity-50"></div>
      <div class="resilver-weather__details text-[4.5cqmin] accent opacity-70 mt-1"></div>
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
    const sym = this._units === "fahrenheit" ? "F" : "C";

    this._iconEl.innerHTML = `<i class="wi wi-${data.icon}"></i>`;
    this._tempEl.innerHTML = `${Math.round(data.temperature)}°<span class="text-[0.75em] opacity-40">${sym}</span>`;
    this._descEl.textContent = data.description;

    const parts = [`Feels like ${Math.round(data.apparentTemperature)}°${sym}`];
    if (this._location) parts.push(this._location);
    this._detailsEl.textContent = parts.join(" · ");
  }
}

customElements.define("resilver-weather", ResilverWeather);
