import React from 'react';
import {
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  SelectChangeEvent,
  Box,
  Typography
} from '@mui/material';

interface NamespaceFilterProps {
  namespaces: string[];
  selectedNamespace: string;
  onNamespaceChange: (namespace: string) => void;
  loading?: boolean;
}

const NamespaceFilter: React.FC<NamespaceFilterProps> = ({
  namespaces,
  selectedNamespace,
  onNamespaceChange,
  loading = false
}) => {
  const handleChange = (event: SelectChangeEvent<string>) => {
    onNamespaceChange(event.target.value);
  };

  return (
    <Box sx={{ minWidth: 200, mb: 2 }}>
      <FormControl fullWidth>
        <InputLabel id="namespace-select-label">Namespace</InputLabel>
        <Select
          labelId="namespace-select-label"
          id="namespace-select"
          value={selectedNamespace}
          label="Namespace"
          onChange={handleChange}
          disabled={loading}
        >
          <MenuItem value="">
            <Typography sx={{ fontStyle: 'italic' }}>All Namespaces</Typography>
          </MenuItem>
          {namespaces.map((namespace) => (
            <MenuItem key={namespace} value={namespace}>
              {namespace}
            </MenuItem>
          ))}
        </Select>
      </FormControl>
    </Box>
  );
};

export default NamespaceFilter;
