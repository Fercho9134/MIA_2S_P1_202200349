import React from 'react';

const Output = ({ data }) => {
  return (
    <div className="p-4 bg-gray-900 text-white h-80 overflow-y-auto">
      <pre>{data}</pre>
    </div>
  );
};

export default Output;
