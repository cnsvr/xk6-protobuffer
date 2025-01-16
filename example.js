import protobuf from 'k6/x/protobuffer';


export default function () {
    const exampleMessage = protobuf.load("example.proto", "ExampleMessage"); 
    exampleMessage.setField('field1', "test1");
    exampleMessage.setField('field2', 123);

    const decoded = exampleMessage.encode();
    console.log(decoded);

    const exampleMessage2 = protobuf.load("example.proto", "ExampleMessage");
    
    exampleMessage2.decode(decoded);
    console.log(exampleMessage2.getField('field1'));
    console.log(exampleMessage2.getField('field2'));
}   
