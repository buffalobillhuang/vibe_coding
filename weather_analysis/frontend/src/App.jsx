import React, { useState, useEffect } from 'react';
import {
  LineChart, Line, BarChart, Bar, XAxis, YAxis,
  CartesianGrid, Tooltip, Legend, ResponsiveContainer
} from 'recharts';

function App() {
  const [yearlyData, setYearlyData] = useState([]);
  const [seasonalData, setSeasonalData] = useState([]);
  const [dailyData, setDailyData] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [selectedView, setSelectedView] = useState('yearly');
  const [year1, setYear1] = useState('2020');
  const [year2, setYear2] = useState('2021');

  // Fetch weather data from API
  useEffect(() => {
    const fetchData = async () => {
      try {
        setLoading(true);
        
        // Fetch yearly data
        const yearlyResponse = await fetch('/api/weather/yearly');
        if (!yearlyResponse.ok) throw new Error('Failed to fetch yearly data');
        const yearlyResult = await yearlyResponse.json();
        setYearlyData(yearlyResult);

        // Fetch seasonal data
        const seasonalResponse = await fetch('/api/weather/seasonal');
        if (!seasonalResponse.ok) throw new Error('Failed to fetch seasonal data');
        const seasonalResult = await seasonalResponse.json();
        setSeasonalData(seasonalResult);

        // Fetch daily data (last year)
        const today = new Date();
        const oneYearAgo = new Date(today.getFullYear() - 1, today.getMonth(), today.getDate());
        const dailyResponse = await fetch(`/api/weather/daily?start_date=${oneYearAgo.toISOString().split('T')[0]}&end_date=${today.toISOString().split('T')[0]}`);
        if (!dailyResponse.ok) throw new Error('Failed to fetch daily data');
        const dailyResult = await dailyResponse.json();
        setDailyData(dailyResult);

        setError(null);
      } catch (err) {
        setError(err.message);
        console.error('Error fetching data:', err);
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, []);

  // Handle comparison data fetch
  const handleCompare = async () => {
    try {
      setLoading(true);
      const response = await fetch(`/api/weather/compare?period=yearly&year1=${year1}&year2=${year2}`);
      if (!response.ok) throw new Error('Failed to fetch comparison data');
      const result = await response.json();
      // Handle comparison data (could update state or show in modal)
      console.log('Comparison data:', result);
      setError(null);
    } catch (err) {
      setError(err.message);
      console.error('Error fetching comparison data:', err);
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return (
      <div className="weather-container">
        <h1>Beijing Weather Analysis</h1>
        <p>Loading weather data...</p>
      </div>
    );
  }

  if (error) {
    return (
      <div className="weather-container">
        <h1>Beijing Weather Analysis</h1>
        <p style={{ color: 'red' }}>Error: {error}</p>
        <p>Using sample data for demonstration...</p>
        {/* Fallback to sample data */}
        <SampleDataView />
      </div>
    );
  }

  return (
    <div className="weather-container">
      <h1>Beijing Weather Analysis (1976-2026)</h1>
      
      {/* Control Panel */}
      <div className="control-panel">
        <select 
          className="select-control"
          value={selectedView} 
          onChange={(e) => setSelectedView(e.target.value)}
        >
          <option value="yearly">Yearly Trends</option>
          <option value="seasonal">Seasonal Trends</option>
          <option value="daily">Daily Trends</option>
        </select>
        
        {selectedView === 'yearly' && (
          <>
            <select 
              className="select-control"
              value={year1} 
              onChange={(e) => setYear1(e.target.value)}
            >
              {[...Array(50)].map((_, i) => (
                <option key={1976 + i} value={1976 + i}>{1976 + i}</option>
              ))}
            </select>
            <select 
              className="select-control"
              value={year2} 
              onChange={(e) => setYear2(e.target.value)}
            >
              {[...Array(50)].map((_, i) => (
                <option key={1976 + i} value={1976 + i}>{1976 + i}</option>
              ))}
            </select>
            <button onClick={handleCompare}>Compare Years</button>
          </>
        )}
      </div>

      {/* Yearly Weather Chart */}
      {selectedView === 'yearly' && (
        <div className="chart-container">
          <h2 className="chart-title">Yearly Temperature and Precipitation Trends</h2>
          <ResponsiveContainer width="100%" height={400}>
            <LineChart data={yearlyData}>
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis dataKey="year" />
              <YAxis yAxisId="left" label={{ value: 'Temperature (°C)', angle: -90, position: 'insideLeft' }} />
              <YAxis yAxisId="right" orientation="right" label={{ value: 'Precipitation (mm)', angle: 90, position: 'insideRight' }} />
              <Tooltip />
              <Legend />
              <Line yAxisId="left" type="monotone" dataKey="avg_temp" stroke="#8884d8" strokeWidth={2} name="Avg. Temperature" />
              <Line yAxisId="right" type="monotone" dataKey="total_precipitation" stroke="#82ca9d" strokeWidth={2} name="Total Precipitation" />
            </LineChart>
          </ResponsiveContainer>
        </div>
      )}

      {/* Seasonal Weather Chart */}
      {selectedView === 'seasonal' && (
        <div className="chart-container">
          <h2 className="chart-title">Seasonal Temperature Trends</h2>
          <ResponsiveContainer width="100%" height={400}>
            <BarChart data={seasonalData}>
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis dataKey="season" />
              <YAxis label={{ value: 'Temperature (°C)', angle: -90, position: 'insideLeft' }} />
              <Tooltip />
              <Legend />
              <Bar dataKey="avg_temp" fill="#8884d8" name="Avg. Temperature" />
            </BarChart>
          </ResponsiveContainer>
        </div>
      )}

      {/* Daily Weather Chart */}
      {selectedView === 'daily' && (
        <div className="chart-container">
          <h2 className="chart-title">Daily Temperature Trends (Last Year)</h2>
          <ResponsiveContainer width="100%" height={400}>
            <LineChart data={dailyData}>
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis dataKey="day" tick={{ fontSize: 10 }} />
              <YAxis label={{ value: 'Temperature (°C)', angle: -90, position: 'insideLeft' }} />
              <Tooltip />
              <Legend />
              <Line type="monotone" dataKey="avg_temp" stroke="#8884d8" strokeWidth={2} name="Avg. Temperature" />
            </LineChart>
          </ResponsiveContainer>
        </div>
      )}
    </div>
  );
}

// Sample data component for demonstration when API fails
function SampleDataView() {
  const sampleYearlyData = [
    { year: '2016', avg_temp: 13.2, total_precipitation: 585 },
    { year: '2017', avg_temp: 13.5, total_precipitation: 610 },
    { year: '2018', avg_temp: 14.0, total_precipitation: 590 },
    { year: '2019', avg_temp: 13.8, total_precipitation: 620 },
    { year: '2020', avg_temp: 14.2, total_precipitation: 630 },
    { year: '2021', avg_temp: 14.5, total_precipitation: 605 },
    { year: '2022', avg_temp: 14.8, total_precipitation: 595 },
    { year: '2023', avg_temp: 15.0, total_precipitation: 615 },
    { year: '2024', avg_temp: 15.2, total_precipitation: 625 },
    { year: '2025', avg_temp: 15.5, total_precipitation: 635 },
  ];

  return (
    <div>
      <div className="chart-container">
        <h2 className="chart-title">Yearly Temperature and Precipitation Trends (Sample Data)</h2>
        <ResponsiveContainer width="100%" height={400}>
          <LineChart data={sampleYearlyData}>
            <CartesianGrid strokeDasharray="3 3" />
            <XAxis dataKey="year" />
            <YAxis yAxisId="left" label={{ value: 'Temperature (°C)', angle: -90, position: 'insideLeft' }} />
            <YAxis yAxisId="right" orientation="right" label={{ value: 'Precipitation (mm)', angle: 90, position: 'insideRight' }} />
            <Tooltip />
            <Legend />
            <Line yAxisId="left" type="monotone" dataKey="avg_temp" stroke="#8884d8" strokeWidth={2} name="Avg. Temperature" />
            <Line yAxisId="right" type="monotone" dataKey="total_precipitation" stroke="#82ca9d" strokeWidth={2} name="Total Precipitation" />
          </LineChart>
        </ResponsiveContainer>
      </div>
      <p style={{ marginTop: '2rem', fontStyle: 'italic' }}>
        Note: This is sample data for demonstration. In a real implementation, data would be fetched from the backend API.
      </p>
    </div>
  );
}

export default App;