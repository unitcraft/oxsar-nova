import { getCLS, getFID, getFCP, getLCP, getTTFB } from 'web-vitals';

interface VitalMetric {
  name: string;
  value: number;
  delta: number;
  id: string;
  rating?: 'good' | 'needs-improvement' | 'poor';
}

export function reportWebVitals(onMetric?: (metric: VitalMetric) => void) {
  // Cumulative Layout Shift — устойчивость макета
  getCLS((metric) => {
    if (onMetric) {
      onMetric({
        name: 'CLS',
        value: metric.value,
        delta: metric.delta,
        id: metric.id,
        rating: metric.rating,
      });
    }
  });

  // First Input Delay — отзывчивость на ввод
  getFID((metric) => {
    if (onMetric) {
      onMetric({
        name: 'FID',
        value: metric.value,
        delta: metric.delta,
        id: metric.id,
        rating: metric.rating,
      });
    }
  });

  // First Contentful Paint — скорость первого контента
  getFCP((metric) => {
    if (onMetric) {
      onMetric({
        name: 'FCP',
        value: metric.value,
        delta: metric.delta,
        id: metric.id,
        rating: metric.rating,
      });
    }
  });

  // Largest Contentful Paint — скорость основного контента
  getLCP((metric) => {
    if (onMetric) {
      onMetric({
        name: 'LCP',
        value: metric.value,
        delta: metric.delta,
        id: metric.id,
        rating: metric.rating,
      });
    }
  });

  // Time to First Byte — скорость ответа сервера
  getTTFB((metric) => {
    if (onMetric) {
      onMetric({
        name: 'TTFB',
        value: metric.value,
        delta: metric.delta,
        id: metric.id,
        rating: metric.rating,
      });
    }
  });
}

// Логирование метрик в консоль (для разработки)
export function logWebVitals(metric: VitalMetric) {
  const label = `%c ${metric.name} `;
  const style = `
    padding: 2px 4px;
    border-radius: 3px;
    color: white;
    font-weight: bold;
    background-color: ${
      metric.rating === 'good'
        ? '#4CAF50'
        : metric.rating === 'needs-improvement'
          ? '#FF9800'
          : '#F44336'
    };
  `;
  console.log(label, style, `${metric.value.toFixed(0)}ms`);
}
