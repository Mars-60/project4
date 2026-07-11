import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import React from "react";
import ReactDOM from "react-dom/client";
import { BrowserRouter, Navigate, Route, Routes } from "react-router-dom";
import { ThemeProvider } from "./lib/theme";
import { AppLayout } from "./ui/AppLayout";
import { ProtectedRoute } from "./ui/ProtectedRoute";
import { Login } from "./pages/Login";
import { Dashboard } from "./pages/Dashboard";
import { Portfolio } from "./pages/Portfolio";
import { Orders } from "./pages/Orders";
import { FundsPage } from "./pages/Funds";
import { PaperTrading } from "./pages/PaperTrading";
import { AIAssistant } from "./pages/AIAssistant";
import { Notifications } from "./pages/Notifications";
import { Settings } from "./pages/Settings";
import { Profile } from "./pages/Profile";
import { StrategyBuilder } from "./pages/StrategyBuilder";
import "./styles.css";

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: 1,
      refetchOnWindowFocus: false
    }
  }
});

ReactDOM.createRoot(document.getElementById("root")!).render(
  <React.StrictMode>
    <ThemeProvider>
      <QueryClientProvider client={queryClient}>
        <BrowserRouter>
          <Routes>
            <Route path="/login" element={<Login />} />
            <Route element={<ProtectedRoute />}>
              <Route element={<AppLayout />}>
                <Route path="/" element={<Dashboard />} />
                <Route path="/strategies" element={<StrategyBuilder />} />
                <Route path="/portfolio" element={<Portfolio />} />
                <Route path="/orders" element={<Orders />} />
                <Route path="/funds" element={<FundsPage />} />
                <Route path="/paper" element={<PaperTrading />} />
                <Route path="/ai" element={<AIAssistant />} />
                <Route path="/notifications" element={<Notifications />} />
                <Route path="/profile" element={<Profile />} />
                <Route path="/settings" element={<Settings />} />
              </Route>
            </Route>
            <Route path="*" element={<Navigate to="/" replace />} />
          </Routes>
        </BrowserRouter>
      </QueryClientProvider>
    </ThemeProvider>
  </React.StrictMode>
);
