import React from 'react';

function App() {
  return (
    <div className="min-h-screen bg-gray-100 flex flex-col items-center justify-center">
      <header className="text-center">
        <h1 className="text-4xl font-bold text-blue-600 mb-4">
          VPS Screener Dashboard
        </h1>
        <p className="text-lg text-gray-700">
          Monitoring your nodes and projects.
        </p>
      </header>
      {/* Placeholder for dashboard content */}
      <main className="mt-8 p-4 w-full max-w-4xl">
        <div className="bg-white shadow-md rounded-lg p-6">
          <h2 className="text-2xl font-semibold text-gray-800 mb-4">System Status</h2>
          <p className="text-gray-600">Node information and project cards will appear here.</p>
          {/* Example of a simple layout that can be expanded */}
          <div className="mt-6 grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            <div className="border border-gray-200 rounded p-4">
              <h3 className="font-semibold text-lg">Node 1 (Alpha)</h3>
              <p className="text-sm text-green-500">Status: Healthy</p>
            </div>
            <div className="border border-gray-200 rounded p-4">
              <h3 className="font-semibold text-lg">Project X on Alpha</h3>
              <p className="text-sm text-gray-500">CPU: 10%, RAM: 200MB</p>
            </div>
            <div className="border border-gray-200 rounded p-4">
              <h3 className="font-semibold text-lg">Node 2 (Beta)</h3>
              <p className="text-sm text-red-500">Status: Unreachable</p>
            </div>
          </div>
        </div>
      </main>
    </div>
  );
}

export default App; 