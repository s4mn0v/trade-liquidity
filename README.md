# Project Structure
```
bitget_liquidity_map/
├── requirements.txt    # List of libraries (streamlit, pandas, plotly, requests)
├── main.py             # The entry point for Streamlit UI
├── core/
│   ├── __init__.py
│   ├── bitget_api.py   # Logic for all GET requests to Bitget V2
│   ├── processor.py    # Logic for Price Binning, Volume Spraying, and Peak Detection
│   └── visualizer.py   # Logic for building the Plotly Figure (Bars + Cumulative Curve)
├── config/
│   ├── __init__.py
│   └── settings.py     # Constants (Colors, Timeframes, API Endpoints, Tick Sizes)
└── assets/
    └── custom_style.css # (Optional) Custom CSS to make Streamlit dark mode prettier
```
