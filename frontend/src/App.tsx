import React from 'react';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import Login from './components/Login';
import Inventory from './components/Inventory';

const App: React.FC = () => {
  return (
    <Router>
      <Routes>
        <Route path="/" element={<Login />} />
        <Route path="/dashboard" element={<Inventory />} />
        {/* Add more routes if needed */}
      </Routes>
    </Router>
  );
};

export default App;