// Copyright TBD

package modbus_methods

/*
   modbus.go - native methods for relish extensions library 'modbus' package. 
*/

import (
   . "relish/runtime/data"
   modbus "relish/runtime/native_methods/extensions/protocols/modbus"
)


func InitModbusMethods() {
   // -------------------------------------------------------------------
   // Modbus 

   // readHoldingRegisters
   //    modbus Modbus
   //    addr Uint32
   //    numRegister Uint32
   // >
   //    pdu Bytes
   readHoldingRegistersMethod, err := RT.CreateMethod("gait.bcit.ca2012/protocols/pkg/modbus",nil,"readHoldingRegisters", []string{"modbus","addr","numRegister"}, []string{"gait.bcit.ca2012/protocols/pkg/modbus/Modbus","Uint32","Uint32"}, []string{"Bytes"}, false, 0, false)
   if err != nil {
      panic(err)
   }
   readHoldingRegistersMethod.PrimitiveCode = readHoldingRegisters

   // writeMultipleRegisters
   //    modbus Modbus
   //    addr Uint32
   //    values []Uint32 // should be []Uint16
   // >
   //    pdu Bytes
   writeMultipleRegistersMethod, err := RT.CreateMethod("gait.bcit.ca2012/protocols/pkg/modbus",nil,"writeMultipleRegisters", []string{"modbus","addr","values"}, []string{"gait.bcit.ca2012/protocols/pkg/modbus/Modbus","Uint32","List_of_Uint32"}, []string{"Bytes"}, false, 0, false)
   if err != nil {
      panic(err)
   }
   writeMultipleRegistersMethod.PrimitiveCode = writeMultipleRegisters

   // writeSingleRegister
   //    modbus Modbus
   //    addr Uint32
   //    value Uint32 // should be Uint16
   // >
   //    pdu Bytes
   writeSingleRegisterMethod, err := RT.CreateMethod("gait.bcit.ca2012/protocols/pkg/modbus",nil,"writeSingleRegister", []string{"modbus","addr","value"}, []string{"gait.bcit.ca2012/protocols/pkg/modbus/Modbus","Uint32","Uint32"}, []string{"Bytes"}, false, 0, false)
   if err != nil {
      panic(err)
   }
   writeSingleRegisterMethod.PrimitiveCode = writeSingleRegister

   // diagnostic
   //    modbus Modbus
   //    value Int32
   // >
   //    pdu Bytes
   diagnosticMethod, err := RT.CreateMethod("gait.bcit.ca2012/protocols/pkg/modbus",nil,"diagnostic", []string{"modbus","value"}, []string{"gait.bcit.ca2012/protocols/pkg/modbus/Modbus","Int32"}, []string{"Bytes"}, false, 0, false)
   if err != nil {
      panic(err)
   }
   diagnosticMethod.PrimitiveCode = diagnostic


   //--------------------------------------------------------------
   // ModbusTCP

   // connect
   //    modbus ModbusTcp
   //    ipAddr String
   //    port Uint32
   //    slaveAddr Uint32
   // >
   //    err String
   connectMethod, err := RT.CreateMethod("gait.bcit.ca2012/protocols/pkg/modbus",nil,"connect", []string{"modbus","ipAddr","port","slaveAddr"}, []string{"gait.bcit.ca2012/protocols/pkg/modbus/ModbusTcp","String","Uint32","Uint32"}, []string{"String"}, false, 0, false)
   if err != nil {
      panic(err)
   }
   connectMethod.PrimitiveCode = connect

   // close
   //    modbus ModbusTcp
   closeMethod, err := RT.CreateMethod("gait.bcit.ca2012/protocols/pkg/modbus",nil,"close", []string{"modbus"}, []string{"gait.bcit.ca2012/protocols/pkg/modbus/ModbusTcp"}, []string{}, false, 0, false)
   if err != nil {
      panic(err)
   }
   closeMethod.PrimitiveCode = close

   // send
   //    modbus Modbus
   //    pdu Bytes
   // >
   //    err String
   sendMethod, err := RT.CreateMethod("gait.bcit.ca2012/protocols/pkg/modbus",nil,"send", []string{"modbus","pdu"}, []string{"gait.bcit.ca2012/protocols/pkg/modbus/ModbusTcp","Bytes"}, []string{"String"}, false, 0, false)
   if err != nil {
      panic(err)
   }
   sendMethod.PrimitiveCode = send

   // read
   //    modbus Modbus
   // >
   //    response Bytes
   //    err String
   readMethod, err := RT.CreateMethod("gait.bcit.ca2012/protocols/pkg/modbus",nil,"read", []string{"modbus"}, []string{"gait.bcit.ca2012/protocols/pkg/modbus/ModbusTcp"}, []string{"Bytes","String"}, false, 0, false)
   if err != nil {
      panic(err)
   }
   readMethod.PrimitiveCode = read



   modbusTcpInitMethod1, err := RT.CreateMethod("gait.bcit.ca2012/protocols/pkg/modbus",nil,"gait.bcit.ca2012/protocols/pkg/modbus/initModbusTcp", []string{"modbus","addressMode","queryTimeout", "queryRetries"}, []string{"gait.bcit.ca2012/protocols/pkg/modbus/ModbusTcp","String","Uint","Uint32"},  []string{"gait.bcit.ca2012/protocols/pkg/modbus/ModbusTcp"}, false, 0, false)
   if err != nil {
      panic(err)
   }
   modbusTcpInitMethod1.PrimitiveCode = initModbusTcp 

}




 
///////////////////////////////////////////////////////////////////////////////////////////
// Modbus functions


// readHoldingRegisters
//    modbus Modbus
//    addr Uint32
//    numRegister Uint32
// >
//    pdu Bytes
func readHoldingRegisters(th InterpreterThread, objects []RObject) []RObject {
   
   wrapper := objects[0].(*GoWrapper)
   addr := uint32(objects[1].(Uint32))
   numRegister := uint32(objects[2].(Uint32))
   modbus := wrapper.GoObj.(modbus.Modbus)
   pdu := modbus.ReadHoldingRegisters(addr,numRegister)
   return []RObject{Bytes(pdu)}
}


// writeMultipleRegisters
//    modbus Modbus
//    addr Uint32
//    values []Uint32
// >
//    pdu Bytes
func writeMultipleRegisters(th InterpreterThread, objects []RObject) []RObject {
   
   wrapper := objects[0].(*GoWrapper)
   addr := uint32(objects[1].(Uint32))
   
   collection := objects[2].(RCollection)
   length := collection.Length()
   values := make([]uint16,length)
   i := 0
   for val := range collection.Iter(th) {
      values[i] = uint16(uint32(val.(Uint32))) // can we convert from Uint32 to uint16?
      i++
   }

   modbus := wrapper.GoObj.(modbus.Modbus)
   pdu := modbus.WriteMultipleRegisters(addr,values)
   return []RObject{Bytes(pdu)}
}


// writeSingleRegisters
//    modbus Modbus
//    addr Uint32
//    value Uint32
// >
//    pdu Bytes
func writeSingleRegister(th InterpreterThread, objects []RObject) []RObject {
   
   wrapper := objects[0].(*GoWrapper)
   addr := uint32(objects[1].(Uint32))
   value := uint16(uint32(objects[2].(Uint32))) // can we convert from Uint32 to uint16?
   modbus := wrapper.GoObj.(modbus.Modbus)
   pdu := modbus.WriteSingleRegister(addr,value)
   return []RObject{Bytes(pdu)}
}


// diagnostic
//    modbus Modbus
//    value Int32
// >
//    pdu Bytes
func diagnostic(th InterpreterThread, objects []RObject) []RObject {
   
   wrapper := objects[0].(*GoWrapper)
   value := int32(objects[1].(Int32))
   modbus := wrapper.GoObj.(modbus.Modbus)
   pdu := modbus.Diagnostic(value)
   return []RObject{Bytes(pdu)}
}


///////////////////////////////////////////////////////////////////////////////////////////
// ModbusTcp functions


// connect
//    modbus ModbusTcp
//    ipAddr String
//    port Uint32
//    slaveAddr Uint32
// >
//    err String
func connect(th InterpreterThread, objects []RObject) []RObject {
   
   wrapper := objects[0].(*GoWrapper)
   ipAddr := string(objects[1].(String))
   port := uint32(objects[2].(Uint32))
   slaveAddr := uint32(objects[3].(Uint32))
   modbus := wrapper.GoObj.(*modbus.ModbusTCP)
   err := modbus.Connect(ipAddr,port,slaveAddr)
   errStr := ""
   if err != nil {
      errStr = err.Error()
   }
   return []RObject{String(errStr)}
}


// close
//    modbus ModbusTcp
func close(th InterpreterThread, objects []RObject) []RObject {
   
   wrapper := objects[0].(*GoWrapper)
   modbus := wrapper.GoObj.(*modbus.ModbusTCP)
   modbus.Close()
   return []RObject{}
}


// send
//    modbus Modbus
//    pdu Bytes
// >
//    err String
func send(th InterpreterThread, objects []RObject) []RObject {
   
   wrapper := objects[0].(*GoWrapper)
   pdu := objects[1].(Bytes)
   p := ([]byte)(pdu)
   modbus := wrapper.GoObj.(*modbus.ModbusTCP)
   err := modbus.Send(p)
   errStr := ""
   if err != nil {
      errStr = err.Error()
   }
   return []RObject{String(errStr)}
}


// read
//    modbus Modbus
// >
//    response Bytes
//    err String
func read(th InterpreterThread, objects []RObject) []RObject {
   
   wrapper := objects[0].(*GoWrapper)
   modbus := wrapper.GoObj.(*modbus.ModbusTCP)
   response, err := modbus.Read()
   errStr := ""
   if err != nil {
      errStr = err.Error()
   }
   return []RObject{Bytes(response),String(errStr)}
}





///////////////////////////////////////////////////////////////////////////////////////////
// Type init functions

func initModbusTcp(th InterpreterThread, objects []RObject) []RObject {
   
   modbusWrapper := objects[0].(*GoWrapper)

   addressMode := string(objects[1].(String))   
   queryTimeout := uint64(objects[2].(Uint))    
   queryRetries := uint32(objects[3].(Uint32))    

   modbusTcp := modbus.MakeModbusTCP( addressMode, queryTimeout, queryRetries )

   modbusWrapper.GoObj = modbusTcp

   return []RObject{modbusWrapper}
}