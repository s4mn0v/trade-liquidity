import streamlit as st
import pandas as pd
import plotly.express as px
import numpy as np

# Simulated price + liquidity data
np.random.seed(42)
prices = np.round(np.linspace(100, 200, 50), 2)
liquidity = np.random.randint(10, 1000, size=len(prices))

df = pd.DataFrame({
    "price": prices,
    "liquidity": liquidity
})

st.title("Liquidity Map (Horizontal Histogram)")

# Create horizontal bar chart
fig = px.bar(
    df,
    x="liquidity",
    y="price",
    orientation="h"
)

fig.update_layout(
    xaxis=dict(fixedrange=True),
    yaxis=dict(fixedrange=True)
)

config = {
    "scrollZoom": False,
    "displayModeBar": False,
    "modeBarButtonsToRemove": [
        "zoom2d",
        "pand2d",
        "select2d",
        "lasso2d",
        "zoomIn2d",
        "zoomOut2d",
        "autoScale2d",
        "resetScale2d"
    ]
}

st.plotly_chart(fig, use_container_width=True, config=config)
