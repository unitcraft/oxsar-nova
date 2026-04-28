interface PerformanceMetric {
  name: string;
  value: number;
  rating?: 'good' | 'needs-improvement' | 'poor';
}

// Отслеживание базовых метрик производительности через Performance API
export function reportWebVitals(onMetric?: (metric: PerformanceMetric) => void) {
  if (typeof window === 'undefined' || !('performance' in window)) {
    return;
  }

  const perf = window.performance;

  // Navigation Timing — время загрузки страницы
  window.addEventListener('load', () => {
    const navTiming = perf.getEntriesByType('navigation')[0] as any;
    if (navTiming) {
      const loadTime = navTiming.loadEventEnd - navTiming.fetchStart;
      if (onMetric) {
        onMetric({
          name: 'Page Load',
          value: loadTime,
          rating: loadTime < 3000 ? 'good' : loadTime < 7500 ? 'needs-improvement' : 'poor',
        });
      }
    }
  });

  // First Contentful Paint (PerformanceObserver)
  if ('PerformanceObserver' in window) {
    try {
      const paintObserver = new PerformanceObserver((list) => {
        for (const entry of list.getEntries()) {
          if (entry.name === 'first-contentful-paint' && onMetric) {
            onMetric({
              name: 'FCP',
              value: entry.startTime,
              rating: entry.startTime < 1800 ? 'good' : entry.startTime < 3000 ? 'needs-improvement' : 'poor',
            });
          }
        }
      });
      paintObserver.observe({ entryTypes: ['paint'] });
    } catch (e) {
      // Ignore if not supported
    }
  }
}

// Логирование метрик в консоль (для разработки)
export function logWebVitals(metric: PerformanceMetric) {
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
