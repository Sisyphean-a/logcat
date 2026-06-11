import React from "react";
import { createRoot } from "react-dom/client";
import App from "./App";
import "./style.css";
import "./log-table.css";
import "./filter-dialog.css";
import "./view-settings.css";

const container = document.getElementById("root");
if (!container) {
  throw new Error("root_not_found");
}

createRoot(container).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>,
);
