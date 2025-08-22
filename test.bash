curl -X POST http://localhost:8080/modificar \
  -H "Content-Type: application/json" \
  -d '[
    {"oldName":"SP002.tif","newName":"SP-20250821-104716-0EC52056-00000666-0000000003000-02705214266596808000041948001-01-04.tif","deleteFile":"SP000.jpg", "moveFile": true},
    {"oldName":"SP003.tif","newName":"SP-20250821-104716-0EC52056-00000666-0000000003000-02705214266596808000041948002-01-04.tif","deleteFile":"SP001.jpg", "moveFile": true},
    {"oldName":"SP006.tif","newName":"SP-20250821-104716-0EC52056-00000666-0000000005000-02705214266596808000041948003-02-04.tif","deleteFile":"SP004.jpg", "moveFile": true},
    {"oldName":"SP007.tif","newName":"SP-20250821-104716-0EC52056-00000666-0000000005000-02705214266596808000041948004-02-04.tif","deleteFile":"SP005.jpg", "moveFile": true},    
    {"oldName":"SP010.tif","newName":"SP-20250821-104716-0EC52056-00000666-0000000013000-02705214266596808000041948005-03-04.tif","deleteFile":"SP008.jpg", "moveFile": true},
    {"oldName":"SP011.tif","newName":"SP-20250821-104716-0EC52056-00000666-0000000013000-02705214266596808000041948006-03-04.tif","deleteFile":"SP009.jpg", "moveFile": true},    
    {"oldName":"SP014.tif","newName":"SP-20250821-104716-0EC52056-00000666-0000000009000-02705214266596808000041948007-04-04.tif","deleteFile":"SP012.jpg", "moveFile": true},
    {"oldName":"SP015.tif","newName":"SP-20250821-104716-0EC52056-00000666-0000000009000-02705214266596808000041948008-04-04.tif","deleteFile":"SP013.jpg", "moveFile": true}
  ]'

