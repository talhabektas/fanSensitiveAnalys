import React from 'react';
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
  ArcElement,
  BarElement,
} from 'chart.js';
import { Line, Pie, Bar, Doughnut } from 'react-chartjs-2';
import { getTeamColors } from '../utils/teamLogos.jsx';

// Register ChartJS components
ChartJS.register(
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
  ArcElement,
  BarElement
);

export const SentimentChart = ({
  data,
  type = 'line',
  height = 300,
  options = {},
  className = '',
}) => {
  // Default options for different chart types
  const defaultOptions = {
    responsive: true,
    maintainAspectRatio: false,
    plugins: {
      legend: {
        position: 'bottom',
        labels: {
          padding: 20,
          usePointStyle: true,
          font: {
            size: 12,
          },
        },
      },
      tooltip: {
        backgroundColor: 'rgba(0, 0, 0, 0.8)',
        titleFont: {
          size: 14,
        },
        bodyFont: {
          size: 12,
        },
        padding: 12,
        cornerRadius: 8,
      },
    },
  };

  // Specific options for different chart types
  const chartOptions = {
    line: {
      ...defaultOptions,
      scales: {
        x: {
          grid: {
            display: false,
          },
          ticks: {
            font: {
              size: 11,
            },
          },
        },
        y: {
          beginAtZero: true,
          grid: {
            color: 'rgba(0, 0, 0, 0.1)',
          },
          ticks: {
            font: {
              size: 11,
            },
          },
        },
      },
      elements: {
        point: {
          radius: 4,
          hoverRadius: 6,
        },
        line: {
          tension: 0.3,
          borderWidth: 2,
        },
      },
    },
    bar: {
      ...defaultOptions,
      scales: {
        x: {
          grid: {
            display: false,
          },
          ticks: {
            font: {
              size: 11,
            },
          },
        },
        y: {
          beginAtZero: true,
          grid: {
            color: 'rgba(0, 0, 0, 0.1)',
          },
          ticks: {
            font: {
              size: 11,
            },
          },
        },
      },
    },
    pie: {
      ...defaultOptions,
      plugins: {
        ...defaultOptions.plugins,
        tooltip: {
          ...defaultOptions.plugins.tooltip,
          callbacks: {
            label: function(context) {
              const label = context.label || '';
              const value = context.parsed;
              const total = context.dataset.data.reduce((sum, val) => sum + val, 0);
              const percentage = total > 0 ? ((value / total) * 100).toFixed(1) : '0.0';
              return `${label}: ${value} (${percentage}%)`;
            },
          },
        },
      },
    },
    doughnut: {
      ...defaultOptions,
      plugins: {
        ...defaultOptions.plugins,
        tooltip: {
          ...defaultOptions.plugins.tooltip,
          callbacks: {
            label: function(context) {
              const label = context.label || '';
              const value = context.parsed;
              const total = context.dataset.data.reduce((sum, val) => sum + val, 0);
              const percentage = total > 0 ? ((value / total) * 100).toFixed(1) : '0.0';
              return `${label}: ${value} (${percentage}%)`;
            },
          },
        },
      },
      cutout: '60%',
    },
  };

  // Merge provided options with defaults
  const finalOptions = {
    ...chartOptions[type],
    ...options,
  };

  // Chart components map
  const ChartComponents = {
    line: Line,
    bar: Bar,
    pie: Pie,
    doughnut: Doughnut,
  };

  const ChartComponent = ChartComponents[type];

  if (!ChartComponent) {
    return (
      <div className={`chart-container ${className}`} style={{ height }}>
        <div className="flex items-center justify-center h-full">
          <p className="text-gray-500">Desteklenmeyen grafik türü: {type}</p>
        </div>
      </div>
    );
  }

  if (!data || !data.datasets || data.datasets.length === 0) {
    return (
      <div className={`chart-container ${className}`} style={{ height }}>
        <div className="flex items-center justify-center h-full">
          <p className="text-gray-500">Grafik verisi bulunamadı</p>
        </div>
      </div>
    );
  }

  return (
    <div className={`chart-container ${className}`} style={{ height }}>
      <ChartComponent data={data} options={finalOptions} />
    </div>
  );
};

// Predefined sentiment chart configurations
export const SentimentDistributionChart = ({ data, height = 300 }) => {
  const chartData = {
    labels: ['Pozitif', 'Nötr', 'Negatif'],
    datasets: [
      {
        data: [
          data?.positive || 0,
          data?.neutral || 0,
          data?.negative || 0,
        ],
        backgroundColor: [
          '#10B981', // Green for positive
          '#6B7280', // Gray for neutral
          '#EF4444', // Red for negative
        ],
        borderColor: [
          '#059669',
          '#4B5563',
          '#DC2626',
        ],
        borderWidth: 2,
        hoverBackgroundColor: [
          '#34D399',
          '#9CA3AF',
          '#F87171',
        ],
      },
    ],
  };

  return (
    <SentimentChart
      data={chartData}
      type="doughnut"
      height={height}
    />
  );
};

export const SentimentTrendChart = ({ data, height = 300 }) => {
  if (!data || data.length === 0) {
    return (
      <div className="chart-container" style={{ height }}>
        <div className="flex items-center justify-center h-full">
          <p className="text-gray-500">Trend verisi bulunamadı</p>
        </div>
      </div>
    );
  }

  const chartData = {
    labels: data.map(item => 
      new Date(item.date).toLocaleDateString('tr-TR', { 
        day: 'numeric', 
        month: 'short' 
      })
    ),
    datasets: [
      {
        label: 'Pozitif',
        data: data.map(item => item.positive),
        borderColor: '#10B981',
        backgroundColor: 'rgba(16, 185, 129, 0.1)',
        fill: true,
      },
      {
        label: 'Negatif',
        data: data.map(item => item.negative),
        borderColor: '#EF4444',
        backgroundColor: 'rgba(239, 68, 68, 0.1)',
        fill: true,
      },
      {
        label: 'Nötr',
        data: data.map(item => item.neutral),
        borderColor: '#6B7280',
        backgroundColor: 'rgba(107, 114, 128, 0.1)',
        fill: true,
      },
    ],
  };

  return (
    <SentimentChart
      data={chartData}
      type="line"
      height={height}
    />
  );
};

export const TeamComparisonChart = ({ data, height = 300 }) => {
  if (!data || data.length === 0) {
    return (
      <div className="chart-container" style={{ height }}>
        <div className="flex items-center justify-center h-full">
          <p className="text-gray-500">Takım karşılaştırma verisi bulunamadı</p>
        </div>
      </div>
    );
  }

  const chartData = {
    labels: data.map(team => team.team_name),
    datasets: [
      {
        label: 'Ortalama Duygu Skoru',
        data: data.map(team => (team.avg_sentiment * 100).toFixed(1)),
        backgroundColor: data.map(team => {
          const teamColors = getTeamColors(team.team_name);
          return teamColors.primary;
        }),
        borderColor: data.map(team => {
          const teamColors = getTeamColors(team.team_name);
          return teamColors.secondary;
        }),
        borderWidth: 2,
      },
    ],
  };

  const options = {
    scales: {
      y: {
        beginAtZero: true,
        max: 100,
        ticks: {
          callback: function(value) {
            return value + '%';
          },
        },
      },
    },
    plugins: {
      tooltip: {
        callbacks: {
          label: function(context) {
            return `${context.dataset.label}: ${context.parsed.y}%`;
          },
        },
      },
    },
  };

  return (
    <SentimentChart
      data={chartData}
      type="bar"
      height={height}
      options={options}
    />
  );
};