import React from 'react';

const TestPage = () => {
  return (
    <div className="p-8">
      <h1 className="text-3xl font-bold text-blue-600 mb-4">
        🎉 Frontend Çalışıyor!
      </h1>
      
      <div className="bg-white rounded-lg shadow-lg p-6 mb-4">
        <h2 className="text-xl font-semibold mb-2">Test Kartı</h2>
        <p className="text-gray-600">
          Bu sayfa görünüyorsa frontend başarıyla çalışıyor demektir.
        </p>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        <div className="bg-gradient-to-r from-purple-500 to-pink-500 text-white p-4 rounded-lg">
          <h3 className="font-bold">Grok AI</h3>
          <p className="text-sm">Ready to go! 🚀</p>
        </div>
        
        <div className="bg-gradient-to-r from-blue-500 to-cyan-500 text-white p-4 rounded-lg">
          <h3 className="font-bold">Backend</h3>
          <p className="text-sm">Port 8060 ✅</p>
        </div>
        
        <div className="bg-gradient-to-r from-green-500 to-teal-500 text-white p-4 rounded-lg">
          <h3 className="font-bold">Frontend</h3>
          <p className="text-sm">Port 3000 ✅</p>
        </div>
      </div>

      <div className="mt-6 p-4 bg-yellow-50 border border-yellow-200 rounded-lg">
        <h4 className="font-semibold text-yellow-800">Sonraki Adımlar:</h4>
        <ul className="list-disc list-inside text-yellow-700 text-sm mt-2">
          <li>AI Dashboard'u test et</li>
          <li>Grok AI entegrasyonunu dene</li>
          <li>Kategori analizini görüntüle</li>
        </ul>
      </div>
    </div>
  );
};

export default TestPage;