import { Navigate, Route, Routes } from "react-router-dom";
import { AppHeader } from "./components/layout/AppHeader";
import { HistoryPage } from "./pages/HistoryPage";
import { HomePage } from "./pages/HomePage";
import { ReportPage } from "./pages/ReportPage";
import { TrainingPage } from "./pages/TrainingPage";

export default function App() {
  return (
    <div className="min-h-screen">
      <AppHeader />
      <main className="animate-fade-in">
        <Routes>
          <Route path="/" element={<HomePage />} />
          <Route path="/training/:sessionId" element={<TrainingPage />} />
          <Route path="/report/:sessionId" element={<ReportPage />} />
          <Route path="/history" element={<HistoryPage />} />
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </main>
    </div>
  );
}
