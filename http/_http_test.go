func TestTail(t *testing.T) {

    lt := NewLogDriverTest("test_tail_file", t)
    filePath := lt.CreateFile("test.txt", "foo\n")

    ld := NewLogDriver("test_tail_file")
    tail, _ := ld.Tail(filePath)

    lt.AppendFile("test.txt", "bar\nbaz\n")
    done := make(chan bool)
    go lt.AssertTailOutput(tail, []string{"foo", "bar", "baz"}, done)
    ld.StopOnReceive(done)
}
