import { Route, Routes } from 'react-router-dom';
import Header from './components/Header.jsx';
import Dashboard from './pages/Dashboard.jsx';
import PlayerProfile from './pages/PlayerProfile.jsx';
import Compare from './pages/Compare.jsx';
import Admin from './pages/Admin.jsx';

export default function App() {
  return (
    <div className="app">
      <Header />
      <main className="app-main">
        <Routes>
          <Route path="/" element={<Dashboard />} />
          <Route path="/players/:id" element={<PlayerProfile />} />
          <Route path="/compare" element={<Compare />} />
          <Route path="/admin" element={<Admin />} />
          <Route path="*" element={<p>Страница не найдена</p>} />
        </Routes>
      </main>
    </div>
  );
}
